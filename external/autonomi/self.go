package autonomi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-autonomi/external/autonomi/models"
)

func (c *Client) GetSelf() (string, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/users/self", c.HostURL), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	self := models.Self{}
	err = json.Unmarshal(resp, &self)
	if err != nil {
		return "", err
	}

	return self.AccountID, nil
}
