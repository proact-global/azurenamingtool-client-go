package azurenamingtool

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetResourceTypes() ([]ResourceTypes, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/v2.0/ResourceTypes", c.HostURL), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	resp := ApiResponse[[]ResourceTypes]{}
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		if resp.Error != nil {
			return nil, fmt.Errorf("[%s] %s", resp.Error.Code, resp.Error.Message)
		}
		return nil, fmt.Errorf("get resource types failed")
	}

	return resp.Data, nil
}
