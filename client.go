package azurenamingtool

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// HostURL - Default API URL
const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	APIKey     string
}

// NewClient -
func NewClient(host, apiKey *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	if apiKey != nil {
		c.APIKey = *apiKey
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.APIKey != "" {
		req.Header.Set("X-API-Key", c.APIKey)
	}

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
