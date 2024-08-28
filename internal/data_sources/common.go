package datasources

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var int64MapAttr = schema.MapAttribute{
	ElementType: types.Int64Type,
	Computed:    true,
}

// combineLocationPairs builds all locations symetrics combinations to be used as Meilisearch filter.
func combineLocationPairs(locations, locationsTo []string) string {
	var result []string

	for _, location := range locations {
		for _, locationTo := range locationsTo {
			result = append(result, fmt.Sprintf("(location = %s && locationTo = %s)", location, locationTo))
			if location != locationTo {
				result = append(result, fmt.Sprintf("(locationTo = %s && location = %s)", location, locationTo))
			}
		}
	}

	return strings.Join(result, " or ")
}

type filter struct {
	Name     types.String `tfsdk:"name"`
	Operator types.String `tfsdk:"operator"`
	Values   types.List   `tfsdk:"values"`
}

type sortFacet struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

type FilterType string

func (ft FilterType) String() string {
	return string(ft)
}

func operatorIntoFilterType(op basetypes.StringValue) FilterType {
	return FilterType(op.ValueString())
}

const (
	EqualFilterType        FilterType = "="
	NotEqualFilterType     FilterType = "!="
	AboveFilterType        FilterType = ">"
	AboveOrEqualFilterType FilterType = ">="
	LessFilterType         FilterType = "<"
	LessOrEqualFilterType  FilterType = "<="
	InFilterType           FilterType = "IN"
	ToFilterType           FilterType = "TO"
)

var (
	ErrWrongOperator = fmt.Errorf(
		"wrong operator, try: \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\"",
		EqualFilterType, NotEqualFilterType, AboveFilterType, AboveOrEqualFilterType, LessFilterType, LessOrEqualFilterType, InFilterType, ToFilterType,
	)
	ErrOnlyOneValue  = errors.New("errors values: must contain one value")
	ErrOnlyTwoValues = errors.New("errors values: must contain two values")
)

// filterTypeValue handles all data to build a Meilisearch filter.
//
// `value`: can be used to store the filter value.
type filterTypeValue struct {
	Name   string
	Type   FilterType
	Values []string
	value  string
}

// build validates and builds the Meilisearch filter.
func (fv *filterTypeValue) build() (string, error) {
	// if the value is already built, skip the build and return the value
	if fv.value != "" {
		return fv.value, nil
	}

	switch fv.Type {
	case ToFilterType:
		if vals := fv.Values; len(vals) == 2 {
			return fmt.Sprintf("%s %s %s %s", fv.Name, vals[0], fv.Type.String(), vals[1]), nil
		}
		return "", ErrOnlyTwoValues
	case InFilterType:
		return fmt.Sprintf("%s %s [%s]", fv.Name, fv.Type.String(), strings.Join(fv.Values[:], ",")), nil
	case EqualFilterType, NotEqualFilterType, AboveFilterType, AboveOrEqualFilterType, LessFilterType, LessOrEqualFilterType:
		if vals := fv.Values; len(vals) == 1 {
			return fmt.Sprintf("%s %s %s", fv.Name, fv.Type.String(), vals[0]), nil
		}
		return "", ErrOnlyOneValue
	default:
		return "", ErrWrongOperator
	}
}

// filters groups all filters data.
type filters map[string]filterTypeValue

func (f filters) addFilter(name string, values []string, type_ FilterType) {
	name = strings.Replace(name, "\"", "", -1)
	f[name] = filterTypeValue{Name: name, Values: values, Type: type_}
}

// setLocationsJunction aims to build a specific locations filter for Transport product.
// If the `location` & `locationTo` are sets, it combines all the location and locationTo pair of values
// to search for symetric locations.
func (f filters) setLocationsJunction() error {
	locationFilter, ok := f["location"]
	if !ok {
		return nil
	}

	locationToFilter, ok := f["locationTo"]
	if !ok {
		return nil
	}

	// validates the filters
	if _, err := locationFilter.build(); err != nil {
		return err
	}

	if _, err := locationToFilter.build(); err != nil {
		return err
	}

	filter := combineLocationPairs(locationFilter.Values, locationToFilter.Values)

	delete(f, "locationTo")
	delete(f, "location")

	f["location"] = filterTypeValue{
		Name:  "locations",
		Type:  EqualFilterType,
		value: filter,
	}

	return nil
}

func (f filters) getFilterList() ([]string, error) {
	var filterValues []string

	if err := f.setLocationsJunction(); err != nil {
		return nil, err
	}

	for _, v := range f {
		filterValue, err := v.build()

		if err != nil {
			return nil, err
		}

		filterValues = append(filterValues, filterValue)
	}

	return filterValues, nil
}

func getFiltersString(fs []filter) ([]string, error) {
	filterStrings := filters{}
	for _, filter := range fs {
		values := []string{}
		for _, v := range filter.Values.Elements() {
			values = append(values, v.String())
		}

		filterStrings.addFilter(filter.Name.String(), values, operatorIntoFilterType(filter.Operator))
	}
	return filterStrings.getFilterList()
}

func getSortString(sorts []sortFacet) []string {
	sortStrings := []string{}
	for _, sort := range sorts {
		sortStrings = append(sortStrings, fmt.Sprintf("%s:%s", sort.Name.ValueString(), sort.Value.ValueString()))
	}
	return sortStrings
}
