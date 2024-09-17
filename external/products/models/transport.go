package models

type TransportProduct struct {
	Product
	LocationTo         string `json:"locationTo"`
	LocationToUnderlay string `json:"locationToUnderlay"`
}

type TransportFacetDistribution struct {
	Bandwidth  map[string]int `json:"bandwidth"`
	Location   map[string]int `json:"location"`
	LocationTo map[string]int `json:"locationTo"`
	Provider   map[string]int `json:"provider"`
}

type TransportProducts struct {
	Hits              []TransportProduct         `json:"hits"`
	FacetDistribution TransportFacetDistribution `json:"facetDistribution"`
}
