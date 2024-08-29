package filters

import (
	"strconv"
	"strings"

	autonomisdkmodel "github.com/intercloud/autonomi-sdk/models"
)

type filterAutonomi struct {
	Name     string
	Values   []string
	Operator FilterType
}

type filtersAutonomi []filterAutonomi

func filtersTerraformToGolang(dataFilters []Filter) (filtersAutonomi, error) {
	var filtersOut filtersAutonomi
	for _, filter := range dataFilters {
		if err := validateFilterAutonomi(filter); err != nil {
			return nil, err
		}
		filtersOut = append(filtersOut, filterTerraformToGolang(&filter))
	}
	return filtersOut, nil
}

func filterTerraformToGolang(dataFilters *Filter) filterAutonomi {
	values := []string{}
	for _, v := range dataFilters.Values.Elements() {
		value := strings.Replace(v.String(), "\"", "", -1)
		values = append(values, value)
	}
	return filterAutonomi{
		Name:     strings.ToLower(dataFilters.Name.ValueString()),
		Operator: operatorIntoFilterType(dataFilters.Operator),
		Values:   values,
	}
}

func validateFilterAutonomi(filter Filter) error {
	switch operatorIntoFilterType(filter.Operator) {
	case EqualFilterType:
		if vals := filter.Values; len(vals.Elements()) == 1 {
			return nil
		}
		return ErrOnlyOneValue
	case InFilterType:
		return nil
	default:
		return ErrWrongOperator
	}
}

func Apply(physicalPorts []autonomisdkmodel.PhysicalPort, filtersTerraform []Filter) ([]autonomisdkmodel.PhysicalPort, error) {
	var result []autonomisdkmodel.PhysicalPort

	filters, err := filtersTerraformToGolang(filtersTerraform)
	if err != nil {
		return nil, err
	}

mainloop:
	for _, physicalPort := range physicalPorts {
		for _, filter := range filters {
			if !filter.match(&physicalPort) {
				continue mainloop
			}
		}
		result = append(result, physicalPort)
	}
	return result, nil
}

func (fg *filterAutonomi) match(physicalPort *autonomisdkmodel.PhysicalPort) bool {
	if fg.Name == "name" {
		return matchesFilter(physicalPort.Name, fg.Operator, fg.Values)
	}
	if fg.Name == "location" {
		return matchesFilter(physicalPort.Product.Location, fg.Operator, fg.Values)
	}
	if fg.Name == "pricemrc" {
		return matchesFilter(strconv.Itoa(physicalPort.Product.PriceMRC), fg.Operator, fg.Values)
	}
	if fg.Name == "pricenrc" {
		return matchesFilter(strconv.Itoa(physicalPort.Product.PriceNRC), fg.Operator, fg.Values)
	}
	return true
}

func matchesFilter(physicalPortValue string, operator FilterType, filterFieldValue []string) bool {
	switch operator {
	case EqualFilterType:
		return physicalPortValue == filterFieldValue[0] // TODO(rmanach): if nil or empty: BOOOOM!
	case InFilterType:
		for _, v := range filterFieldValue {
			if physicalPortValue == v {
				return true
			}
		}
		return false
	}
	return false
}
