package filters

import (
	"fmt"
	"strings"
)

// filterMeilisearch handles all data to build a Meilisearch filter.
// `value`: can be used to store the filter value.
type filterMeilisearch struct {
	Name   string
	Type   FilterType
	Values []string
	value  string
}

// filters groups all filters data.
type filterMeilisearchsMap map[string]filterMeilisearch

// build validates and builds the Meilisearch filter.
func (fv *filterMeilisearch) build() (string, error) {
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

func (f filterMeilisearchsMap) addFilter(name string, values []string, type_ FilterType) {
	name = strings.Replace(name, "\"", "", -1)
	f[name] = filterMeilisearch{Name: name, Values: values, Type: type_}
}

// setLocationsJunction aims to build a specific locations filter for Transport product.
// If the `location` & `locationTo` are sets, it combines all the location and locationTo pair of values
// to search for symetric locations.
func (f filterMeilisearchsMap) setLocationsJunction() error {
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

	f["location"] = filterMeilisearch{
		Name:  "locations",
		Type:  EqualFilterType,
		value: filter,
	}

	return nil
}

func (f filterMeilisearchsMap) getFilterList() ([]string, error) {
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
