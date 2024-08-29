package filters

import (
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

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

var Int64MapAttr = schema.MapAttribute{
	ElementType: types.Int64Type,
	Computed:    true,
}

var (
	ErrWrongOperator = fmt.Errorf(
		"wrong operator, try: \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\"",
		EqualFilterType, NotEqualFilterType, AboveFilterType, AboveOrEqualFilterType, LessFilterType, LessOrEqualFilterType, InFilterType, ToFilterType,
	)
	ErrOnlyOneValue  = errors.New("errors values: must contain one value")
	ErrOnlyTwoValues = errors.New("errors values: must contain two values")
)

type Filter struct {
	Name     types.String `tfsdk:"name"`
	Operator types.String `tfsdk:"operator"`
	Values   types.List   `tfsdk:"values"`
}

type FilterType string

type SortFacet struct {
	Name  types.String `tfsdk:"name"`
	Value types.String `tfsdk:"value"`
}

func (ft FilterType) String() string {
	return string(ft)
}

func operatorIntoFilterType(op basetypes.StringValue) FilterType {
	return FilterType(op.ValueString())
}

func GetFiltersString(fs []Filter) ([]string, error) {
	filterStrings := filterMeilisearchsMap{}
	for _, filter := range fs {
		values := []string{}
		for _, v := range filter.Values.Elements() {
			values = append(values, v.String())
		}

		filterStrings.addFilter(filter.Name.String(), values, operatorIntoFilterType(filter.Operator))
	}
	return filterStrings.getFilterList()
}

func GetSortString(sorts []SortFacet) []string {
	sortStrings := []string{}
	for _, sort := range sorts {
		sortStrings = append(sortStrings, fmt.Sprintf("%s:%s", sort.Name.ValueString(), sort.Value.ValueString()))
	}
	return sortStrings
}
