package azurenamingtool

import (
    "fmt"
    "net/http"
	"strings"
)

func (c *Client) DeleteName(deletename DeleteGeneratedNameRequest) ([]byte, error) {
	// Admin endpoints use the V1 path (no version prefix).
	req, err := http.NewRequest("DELETE",
		fmt.Sprintf("%s/api/Admin/DeleteGeneratedName/%d", c.HostURL, deletename.ID),
		nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.mu.Lock()
	body, err := c.doRequest(req)
	c.mu.Unlock()
	if err != nil {
		// Treat "not found" as success — the entry is already gone.
		if strings.Contains(err.Error(), "Generated Name not found") {
			return nil, nil
		}
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// The V1 Admin API always returns HTTP 200; failures are indicated in the body.
	bodyStr := strings.TrimSpace(string(body))
	if strings.Contains(bodyStr, "FAILURE") {
		return nil, fmt.Errorf("delete failed: %s", bodyStr)
	}

	return body, nil
}
