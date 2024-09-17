package models

type CloudProduct struct {
	Product
	CSPName string `json:"cspName"`
}

type CloudFacetDistribution struct {
	Bandwidth map[string]int `json:"bandwidth"`
	CSPCity   map[string]int `json:"cspCity"`
	CSPName   map[string]int `json:"cspName"`
	CSPRegion map[string]int `json:"cspRegion"`
	Location  map[string]int `json:"location"`
	Provider  map[string]int `json:"provider"`
}

type CloudProducts struct {
	Hits              []CloudProduct         `json:"hits"`
	FacetDistribution CloudFacetDistribution `json:"facetDistribution"`
}
