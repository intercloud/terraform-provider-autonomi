package models

type AccessType string

const (
	PHYSICAL AccessType = "PHYSICAL"
	VIRTUAL  AccessType = "VIRTUAL"
)

func (at AccessType) String() string {
	return string(at)
}

type AccessProduct struct {
	Product
	Type string `json:"type"`
}

type AccessFacetDistribution struct {
	Bandwidth map[string]int `json:"bandwidth"`
	Location  map[string]int `json:"location"`
	Provider  map[string]int `json:"provider"`
	Type      map[string]int `json:"type"`
}

type AccessProducts struct {
	Hits              []AccessProduct         `json:"hits"`
	FacetDistribution AccessFacetDistribution `json:"facetDistribution"`
}
