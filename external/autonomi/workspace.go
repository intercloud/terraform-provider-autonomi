package autonomi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"terraform-provider-autonomi/external/autonomi/models"
)

func (c *Client) CreateWorkspace(payload models.CreateWorkspace) (*models.Workspace, error) {
	body := new(bytes.Buffer)
	err := json.NewEncoder(body).Encode(&payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/accounts/%s/workspaces", c.HostURL, c.AccountID), body)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	workspace := models.WorkspaceResponse{}
	if err := json.Unmarshal(resp, &workspace); err != nil {
		return nil, err
	}

	return &workspace.Data, nil
}

func (c *Client) GetWorkspace(workspaceID string) (*models.Workspace, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/accounts/%s/workspaces/%s", c.HostURL, c.AccountID, workspaceID), nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	workspace := models.WorkspaceResponse{}
	err = json.Unmarshal(resp, &workspace)
	if err != nil {
		return nil, err
	}

	return &workspace.Data, nil
}

func (c *Client) DeleteWorkspace(workspaceID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/accounts/%s/workspaces/%s", c.HostURL, c.AccountID, workspaceID), nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
