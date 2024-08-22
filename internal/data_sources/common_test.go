package datasources

import (
	"errors"
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
		filters []filter
		expect  []string
		err     error
	}{
		{
			name: EqualFilterType.String(),
			filters: []filter{
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
			filters: []filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{"100", "200"}),
				},
			},
			expect: nil,
			err:    errors.New("errors values: must contain one value"),
		},
		{
			name: EqualFilterType.String() + " must fail len(values) < 1",
			filters: []filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(EqualFilterType.String()),
					Values:   getValues([]string{}),
				},
			},
			expect: nil,
			err:    errors.New("errors values: must contain one value"),
		},
		{
			name: ToFilterType.String(),
			filters: []filter{
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
			filters: []filter{
				{
					Name:     types.StringValue("bandwidth"),
					Operator: types.StringValue(ToFilterType.String()),
					Values:   getValues([]string{"100"}),
				},
			},
			expect: nil,
			err:    errors.New("errors values: must contain two values"),
		},
		{
			name: InFilterType.String(),
			filters: []filter{
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
			name: InFilterType.String() + " empty list",
			filters: []filter{
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
			filters: []filter{
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
		filterString, err := getFiltersString(tc.filters)
		assert.Equal(t, tc.err, err)
		assert.Equal(t, tc.expect, filterString)
	}
}
