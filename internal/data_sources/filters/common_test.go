package filters

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/stretchr/testify/assert"
)

func getValues(listValues []string) basetypes.ListValue {
	attr := []attr.Value{}
	for _, v := range listValues {
		attr = append(attr, types.StringValue(v))
	}
	values, err := types.ListValue(
		types.StringType,
		attr,
	)
	if err != nil {
		panic(err)
	}
	return values
}

func TestGetFiltersString(t *testing.T) {
	tests := []struct {
		name    string
		filters []Filter
		expect  []string
		err     error
	}{
		{
			name: EqualFilterType.String(),
			filters: []Filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"100"}),
				},
			},
			expect: []string{"bandwidth = \"100\""},
			err:    nil,
		},
		{
			name: EqualFilterType.String() + " must fail len(values) > 1",
			filters: []Filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"100", "200"}),
				},
			},
			expect: nil,
			err:    ErrOnlyOneValue,
		},
		{
			name: EqualFilterType.String() + " must fail len(values) < 1",
			filters: []Filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{}),
				},
			},
			expect: nil,
			err:    ErrOnlyOneValue,
		},
		{
			name: ToFilterType.String(),
			filters: []Filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(ToFilterType.String()),
					Values:   getValues([]string{"100", "200"}),
				},
			},
			expect: []string{"bandwidth \"100\" TO \"200\""},
			err:    nil,
		},
		{
			name: ToFilterType.String() + " must fail len(values) != 2",
			filters: []Filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(ToFilterType.String()),
					Values:   getValues([]string{"100"}),
				},
			},
			expect: nil,
			err:    ErrOnlyTwoValues,
		},
		{
			name: InFilterType.String(),
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{"EQUINIX FR5", "EQUINIX LD5"}),
				},
			},
			expect: []string{"location IN [\"EQUINIX FR5\",\"EQUINIX LD5\"]"},
			err:    nil,
		},
		{
			name: "transport junction",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"EQUINIX FR5"}),
				},
				{
					Name:     types.StringValue("locationTo"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"EQUINIX LD5"}),
				},
			},
			expect: []string{"(location = \"EQUINIX FR5\" AND locationTo = \"EQUINIX LD5\") OR (locationTo = \"EQUINIX FR5\" AND location = \"EQUINIX LD5\")"},
			err:    nil,
		},
		{
			name: "transport junction different filter type",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{"EQUINIX FR5"}),
				},
				{
					Name:     types.StringValue("locationTo"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"EQUINIX LD5"}),
				},
			},
			expect: []string{"(location = \"EQUINIX FR5\" AND locationTo = \"EQUINIX LD5\") OR (locationTo = \"EQUINIX FR5\" AND location = \"EQUINIX LD5\")"},
			err:    nil,
		},
		{
			name: "transport junction with multiple locations bad operator",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{"EQUINIX FR5", "EQUINIX AM2", "EQUINIX SG1"}),
				},
				{
					Name:     types.StringValue("locationTo"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"EQUINIX AM2", "EQUINIX PA3"}),
				},
			},
			expect: nil,
			err:    ErrOnlyOneValue,
		},
		{
			name: "transport junction with multiple locations",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{"EQUINIX FR5", "EQUINIX AM2", "EQUINIX SG1"}),
				},
				{
					Name:     types.StringValue("locationTo"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{"EQUINIX AM2", "EQUINIX PA3"}),
				},
			},
			expect: []string{"(location = \"EQUINIX FR5\" AND locationTo = \"EQUINIX AM2\") OR (locationTo = \"EQUINIX FR5\" AND location = \"EQUINIX AM2\") OR (location = \"EQUINIX FR5\" AND locationTo = \"EQUINIX PA3\") OR (locationTo = \"EQUINIX FR5\" AND location = \"EQUINIX PA3\") OR (location = \"EQUINIX AM2\" AND locationTo = \"EQUINIX AM2\") OR (location = \"EQUINIX AM2\" AND locationTo = \"EQUINIX PA3\") OR (locationTo = \"EQUINIX AM2\" AND location = \"EQUINIX PA3\") OR (location = \"EQUINIX SG1\" AND locationTo = \"EQUINIX AM2\") OR (locationTo = \"EQUINIX SG1\" AND location = \"EQUINIX AM2\") OR (location = \"EQUINIX SG1\" AND locationTo = \"EQUINIX PA3\") OR (locationTo = \"EQUINIX SG1\" AND location = \"EQUINIX PA3\")"},
			err:    nil,
		},
		{
			name: InFilterType.String() + " empty list",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue(InFilterType.String()),
					Values:   getValues([]string{}),
				},
			},
			expect: []string{"location IN []"},
			err:    nil,
		},
		{
			name: "wrong operator",
			filters: []Filter{
				{
					Name:     types.StringValue("location"),
					Operator: types.StringValue("plouf"),
					Values:   getValues([]string{}),
				},
			},
			expect: nil,
			err:    ErrWrongOperator,
		},
	}

	for _, tc := range tests {
		t.Log(tc.name)
		tc := tc
		filterString, err := GetFiltersString(tc.filters)
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.expect, filterString)
	}
}
