package models

type CloudFilters struct {
	CSPName   string
	CSPRegion string
	CSPCity   string
	Provider  string
	Location  string
	Bandwidth int
}

type CloudProduct struct {
	ID        int    `json:"id"`
	Provider  string `json:"provider"`
	Duration  int    `json:"duration"`
	Location  string `json:"location"`
	Bandwidth int    `json:"bandwidth"`
	Date      string `json:"date"`
	PriceNRC  int    `json:"priceNrc"`
	PriceMRC  int    `json:"priceMrc"`
	CostNRC   int    `json:"costNrc"`
	CostMRC   int    `json:"costMrc"`
	SKU       string `json:"sku"`
	CSPName   string `json:"cspName"`
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
