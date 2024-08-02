package autonomi

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HostURL - Default Autonomi URL
const (
	HostURL string = "https://api-platform-dev.intercloud.io/v1"
)

var timeout = 30 * time.Second

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
	AccountID  string
}

// NewClient -
func NewClient(personalAccessToken string, termsAndConditions bool) (*Client, error) {
	if !termsAndConditions {
		return nil, errors.New("terms and conditions must be accepted")
	}
	c := Client{
		HTTPClient: &http.Client{Timeout: timeout},
		// Default Autonomi URL
		HostURL: HostURL,
		Token:   personalAccessToken,
	}

	accountID, err := c.GetSelf()
	if err != nil {
		return nil, err
	}

	c.AccountID = accountID

	return &c, nil
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

	if res.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
