package datasources

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var int64MapAttr = schema.MapAttribute{
	ElementType: types.Int64Type,
	Computed:    true,
}

type filter struct {
	Name     types.String `tfsdk:"name"`
	Operator types.String `tfsdk:"operator"`
	Values   types.List   `tfsdk:"values"`
}

type FilterType string

func (ft FilterType) String() string {
	return string(ft)
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

var ErrWrongOperator = fmt.Errorf(
	"wrong operator, try: \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\", \"%s\"",
	EqualFilterType, NotEqualFilterType, AboveFilterType, AboveOrEqualFilterType, LessFilterType, LessOrEqualFilterType, InFilterType, ToFilterType,
)

func getFiltersString(filters []filter) ([]string, error) {
	filterStrings := []string{}
	for _, filter := range filters {
		switch filter.Operator.ValueString() {
		case ToFilterType.String():
			if len(filter.Values.Elements()) == 2 {
				value1 := filter.Values.Elements()[0].String()
				value2 := filter.Values.Elements()[1].String()
				filterStrings = append(filterStrings, fmt.Sprintf("%s %s %s %s", filter.Name.ValueString(), value1, filter.Operator.ValueString(), value2))
			} else {
				return nil, errors.New("errors values: must contain two values")
			}
		case InFilterType.String():
			values := []string{}
			for _, v := range filter.Values.Elements() {
				values = append(values, v.String())
			}
			filterStrings = append(filterStrings, fmt.Sprintf("%s %s [%s]", filter.Name.ValueString(), filter.Operator.ValueString(), strings.Join(values[:], ",")))
		case EqualFilterType.String(), NotEqualFilterType.String(), AboveFilterType.String(), AboveOrEqualFilterType.String(), LessFilterType.String(), LessOrEqualFilterType.String():
			if len(filter.Values.Elements()) == 1 {
				value := filter.Values.Elements()[0].String()
				filterStrings = append(filterStrings, fmt.Sprintf("%s %s %s", filter.Name.ValueString(), filter.Operator.ValueString(), value))
			} else {
				return nil, errors.New("errors values: must contain one value")
			}
		default:
			return nil, ErrWrongOperator
		}
	}
	return filterStrings, nil
}
