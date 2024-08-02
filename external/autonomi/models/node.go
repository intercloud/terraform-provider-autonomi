package models

import (
	"time"
)

type Product struct {
	Provider  string    `json:"provider"`
	Duration  int       `json:"duration"`
	Location  string    `json:"location"`
	Bandwidth int       `json:"bandwidth"`
	Date      time.Time `json:"date"`
	PriceNRC  int       `json:"priceNrc"`
	PriceMRC  int       `json:"priceMrc"`
	CostNRC   int       `json:"costNrc"`
	CostMRC   int       `json:"costMrc"`
	SKU       string    `json:"sku"`
}

type NodeProduct struct {
	Product
	CSPName         string `json:"cspName,omitempty"`
	CSPNameUnderlay string `json:"cspNameUnderlay,omitempty"`
	CSPCity         string `json:"cspCity,omitempty"`
	CSPRegion       string `json:"cspRegion,omitempty"`
}

type Port struct {
	ID              string `json:"id"`
	LocationID      string `json:"locationId"`
	CSPName         string `json:"cspName"`
	CSPNameUnderlay string `json:"cspNameUnderlay"`
}

type ProviderCloudConfig struct {
	PairingKey string `json:"pairingKey,omitempty"`
	AccountID  string `json:"accountId,omitempty"`
	ServiceKey string `json:"serviceKey,omitempty"`
}

type Node struct {
	BaseModel
	AccountID      string               `json:"accountId"`
	WorkspaceID    string               `json:"workspaceId"`
	Name           string               `json:"name"`
	State          string               `json:"administrativeState"`
	DeployedAt     *time.Time           `json:"deployedAt,omitempty"`
	Product        NodeProduct          `json:"product,omitempty"`
	Type           string               `json:"type,omitempty"`
	ConnectionID   string               `json:"connectionId,omitempty"`
	Port           *Port                `json:"port,omitempty"`
	ProviderConfig *ProviderCloudConfig `json:"providerConfig,omitempty"`
	Vlan           int64                `json:"vlan,omitempty"`
	DxconID        string               `json:"dxconId,omitempty"`
}

type NodeResponse struct {
	Data Node `json:"data"`
}

type CreateNode struct {
	WorkspaceID    string               `json:"workspaceId"`
	Name           string               `json:"name"`
	Type           string               `json:"type"`
	Product        NodeProduct          `json:"product"`
	ProviderConfig *ProviderCloudConfig `json:"providerConfig"`
}
