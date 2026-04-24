package azurenamingtool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

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

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	resp := ApiResponse[ResourceNameDetails]{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		if resp.Error != nil {
			return nil, fmt.Errorf("[%s] %s", resp.Error.Code, resp.Error.Message)
		}
		return nil, fmt.Errorf("get name failed")
	}

	return &resp.Data, nil
}