package models

type PhysicalPortProduct Product

type PhysicalPortProductFacetDistribution struct {
	Bandwidth map[string]int `json:"bandwidth"`
	Location  map[string]int `json:"location"`
	Provider  map[string]int `json:"provider"`
	Duration  map[string]int `json:"duration"`
}

type PhysicalPortProducts struct {
	Hits              []PhysicalPortProduct                `json:"hits"`
	FacetDistribution PhysicalPortProductFacetDistribution `json:"facetDistribution"`
}
