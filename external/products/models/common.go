package models

import (
	autonomisdk "github.com/intercloud/autonomi-sdk"
	"github.com/meilisearch/meilisearch-go"
)

type ProviderType string

const (
	INTERCLOUD ProviderType = "InterCloud"
	EQUINIX    ProviderType = "EQUINIX"
	MEGAPORT   ProviderType = "MEGAPORT"
)

func (pt ProviderType) String() string {
	return string(pt)
}

type Product struct {
	ID               int    `json:"id"`
	Provider         string `json:"provider"`
	Duration         int    `json:"duration"`
	Location         string `json:"location"`
	LocationUnderlay string `json:"locationUnderlay"`
	Bandwidth        int    `json:"bandwidth"`
	Date             string `json:"date"`
	PriceNRC         int    `json:"priceNrc"`
	PriceMRC         int    `json:"priceMrc"`
	CostNRC          int    `json:"costNrc"`
	CostMRC          int    `json:"costMrc"`
	SKU              string `json:"sku"`
}

type Clients struct {
	CatalogClient  *meilisearch.Client
	AutonomiClient *autonomisdk.Client
}
