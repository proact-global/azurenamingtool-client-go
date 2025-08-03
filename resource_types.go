package azurenamingtool

import (
	"encoding/json"
	"fmt"
	"net/http"
	//"strings"
)

// GetCoffees - Returns list of coffees (no auth required)
func (c *Client) GetResourceTypes() ([]ResourceTypes, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/ResourceTypes", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, nil)
	if err != nil {
		return nil, err
	}

	resourcetypes := []ResourceTypes{}
	err = json.Unmarshal(body, &resourcetypes)
	if err != nil {
		return nil, err
	}

	return resourcetypes, nil
}