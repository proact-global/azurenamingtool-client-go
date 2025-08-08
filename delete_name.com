package azurenamingtool

import (
    "fmt"
    "net/http"
)

func (c *Client) DeleteName(deletename DeleteGeneratedNameRequest) ([]byte, error) {
    req, err := http.NewRequest("DELETE",
        fmt.Sprintf("%s/api/Admin/DeleteGeneratedName/%d", c.HostURL, deletename.ID),
        nil)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }

    body, err := c.doRequest(req)
    if err != nil {
        return nil, fmt.Errorf("request failed: %w", err)
    }

    return body, nil
}
