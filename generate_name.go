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

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/ResourceNamingRequests/RequestName", c.HostURL), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	newName := GenerateNameResponse{}
	err = json.Unmarshal(body, &newName)
	if err != nil {
		return nil, err
	}

	return &newName, nil
}

// GetName - Returns a Name on ID
func (c *Client) GetName(NameID int16) (*ResourceNameDetails, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/Admin/GetGeneratedName/%d", c.HostURL, NameID), nil)
    if err != nil {
        return nil, err
    }

    body, err := c.doRequest(req)
    if err != nil {
        return nil, err
    }

	Name := ResourceNameDetails{}
	err = json.Unmarshal(body, &Name)
	if err != nil {
		return nil, err
	}

	return &Name, nil
}