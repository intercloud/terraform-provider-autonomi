package products

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// HostURL - Default Catalog URL
const HostURL string = "https://search-platform-dev.intercloud.io/indexes/cloudproduct/search"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

type cloudProduct struct {
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

type products struct {
	Hits []cloudProduct `json:"hits"`
}

type productDataSourceRequest struct {
	Filter string   `json:"filter"`
	Facets []string `json:"facets"`
}

func NewClient(personalAccessToken string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		// Default Catalog URL
		HostURL: HostURL,
		Token:   personalAccessToken,
	}
	return &c, nil
}

func createFilter(cspName, provider, location, bandwidth string) string {
	var filters []string

	if cspName != "" {
		filters = append(filters, fmt.Sprintf(`cspName = "%s"`, cspName))
	}
	if provider != "" {
		filters = append(filters, fmt.Sprintf(`provider = "%s"`, provider))
	}
	if location != "" {
		filters = append(filters, fmt.Sprintf(`location = "%s"`, location))
	}
	if bandwidth != "" {
		filters = append(filters, fmt.Sprintf(`bandwidth = "%s"`, bandwidth))
	}

	return strings.Join(filters, " AND ")
}

func (c *Client) GetCloudProducts() ([]cloudProduct, error) {

	payload := productDataSourceRequest{
		Filter: `cspName = "AWS" AND provider = "EQUINIX" AND location = "EQUINIX LD5" AND bandwidth = "100"`,
		Facets: []string{"cspName", "cspRegion", "cspCity", "location", "bandwidth", "provider"},
	}

	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", c.HostURL, body)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	products := products{}
	err = json.Unmarshal(resp, &products)
	if err != nil {
		return nil, err
	}

	return products.Hits, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Add("Content-Type", "application/json")

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
