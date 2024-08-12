package models

type TransportFilters struct {
	Provider   string
	Location   string
	LocationTo string
	Bandwidth  int
}

type TransportProduct struct {
	ID                 int    `json:"id"`
	Provider           string `json:"provider"`
	Location           string `json:"location"`
	LocationUnderlay   string `json:"locationUnderlay"`
	LocationTo         string `json:"locationTo"`
	LocationToUnderlay string `json:"locationToUnderlay"`
	Bandwidth          int    `json:"bandwidth"`
	Date               string `json:"date"`
	PriceNRC           int    `json:"priceNrc"`
	PriceMRC           int    `json:"priceMrc"`
	CostNRC            int    `json:"costNrc"`
	CostMRC            int    `json:"costMrc"`
	SKU                string `json:"sku"`
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
