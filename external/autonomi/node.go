package autonomi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-autonomi/external/autonomi/models"
)

func (c *Client) CreateNode(payload models.CreateNode, workspaceID string) (*models.Node, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/accounts/%s/workspaces/%s/nodes", c.HostURL, c.AccountID, workspaceID), body)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	node := models.NodeResponse{}
	err = json.Unmarshal(resp, &node)
	if err != nil {
		return nil, err
	}

	return &node.Data, err
}

func (c *Client) GetNode(workspaceID, nodeID string) (*models.Node, error) {

	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/accounts/%s/workspaces/%s/nodes/%s", c.HostURL, c.AccountID, workspaceID, nodeID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	node := models.NodeResponse{}
	err = json.Unmarshal(resp, &node)
	if err != nil {
		return nil, err
	}

	return &node.Data, err
}

func (c *Client) DeleteNode(workspaceID, nodeID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/accounts/%s/workspaces/%s/nodes/%s", c.HostURL, c.AccountID, workspaceID, nodeID), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
