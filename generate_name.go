package azurenamingtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrNotFound is returned by GetName when the requested ID does not exist in the naming tool.
var ErrNotFound = errors.New("generated name not found")

// GenerateName - Generate new name
func (c *Client) GenerateName(generatename GenerateNameRequest) (*GenerateNameResponse, error) {
	rb, err := json.Marshal(generatename)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2.0/ResourceNamingRequests/RequestName", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	body, err := c.doRequest(req)
	c.mu.Unlock()
	if err != nil {
		return nil, err
	}

	resp := ApiResponse[GenerateNameResponse]{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		if resp.Error != nil {
			return nil, fmt.Errorf("[%s] %s", resp.Error.Code, resp.Error.Message)
		}
		return nil, fmt.Errorf("name generation failed")
	}

	return &resp.Data, nil
}

// GetName - Returns a Name on ID
func (c *Client) GetName(NameID int64) (*ResourceNameDetails, error) {
	// Admin endpoints use the V1 path (no version prefix).
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/Admin/GetGeneratedName/%d", c.HostURL, NameID), nil)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	body, err := c.doRequest(req)
	c.mu.Unlock()
	if err != nil {
		// The Admin endpoint returns HTTP 400 with a "not found" message when the
		// ID does not exist. Map that to the ErrNotFound sentinel so callers can
		// distinguish a missing resource from a transient failure.
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return nil, ErrNotFound
		}
		return nil, err
	}

	// The Admin endpoint returns the raw GeneratedName object directly,
	// not wrapped in the V2 ApiResponse envelope.
	resp := ResourceNameDetails{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	return &resp, nil
}
