package slicer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// SlicerClient handles all HTTP communication with the Slicer API
type SlicerClient struct {
	httpClient *http.Client
	baseURL    string
	token      string
	userAgent  string
}

// NewSlicerClient creates a new Slicer API client
func NewSlicerClient(baseURL, token string, userAgent string, httpClient *http.Client) *SlicerClient {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &SlicerClient{
		httpClient: httpClient,
		baseURL:    baseURL,
		token:      token,
		userAgent:  userAgent,
	}
}

// makeRequest creates and executes an HTTP request with proper authentication
func (c *SlicerClient) makeRequest(method, endpoint string, body interface{}) (*http.Response, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	u.Path = path.Join(u.Path, endpoint)

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequest(method, u.String(), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if c.userAgent != "" {
		req.Header.Set("User-Agent", c.userAgent)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return c.httpClient.Do(req)
}

// GetHostGroups fetches all host groups from the API
func (c *SlicerClient) GetHostGroups() ([]SlicerHostGroup, error) {
	res, err := c.makeRequest(http.MethodGet, "/hostgroup", nil)
	if err != nil {
		return nil, err
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s - %s", res.Status, string(body))
	}

	var hostGroups []SlicerHostGroup
	if err := json.Unmarshal(body, &hostGroups); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return hostGroups, nil
}

// GetHostGroupNodes fetches nodes for a specific host group
func (c *SlicerClient) GetHostGroupNodes(groupName string) ([]SlicerNode, error) {
	endpoint := fmt.Sprintf("hostgroup/%s/nodes", groupName)
	res, err := c.makeRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch nodes: %w", err)
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed: %s - %s", res.Status, string(body))
	}

	var nodes []SlicerNode
	if err := json.Unmarshal(body, &nodes); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return nodes, nil
}

// CreateNode creates a new node in the specified host group
func (c *SlicerClient) CreateNode(groupName string, request SlicerCreateNodeRequest) (*SlicerCreateNodeResponse, error) {
	endpoint := fmt.Sprintf("hostgroup/%s/nodes", groupName)
	res, err := c.makeRequest(http.MethodPost, endpoint, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("API request failed: %s - %s", res.Status, string(body))
	}

	var result SlicerCreateNodeResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &result, nil
}

// DeleteNode deletes a node from the specified host group
func (c *SlicerClient) DeleteNode(groupName, nodeName string) error {
	endpoint := fmt.Sprintf("hostgroup/%s/nodes/%s", groupName, nodeName)
	res, err := c.makeRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to delete node: %w", err)
	}

	var body []byte
	if res.Body != nil {
		defer res.Body.Close()
		body, _ = io.ReadAll(res.Body)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("API request failed: %s - %s", res.Status, string(body))
	}

	return nil
}
