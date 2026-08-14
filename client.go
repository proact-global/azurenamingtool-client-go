package azurenamingtool

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

// HostURL - Default API URL
const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HostURL       string
	HTTPClient    *http.Client
	APIKey        string
	AdminPassword *string
	mu            sync.Mutex
}

// NewClient -
func NewClient(host, apiKey, AdminPassword *string) (*Client, error) {
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

	if AdminPassword != nil {
		c.AdminPassword = AdminPassword
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	if c.APIKey != "" {
		req.Header.Set("APIKey", c.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "*/*")

	if c.AdminPassword != nil {
		req.Header.Set("AdminPassword", *c.AdminPassword)
	}

	// Perform the request

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
		apiErr := &APIError{
			StatusCode: res.StatusCode,
			Body:       string(body),
		}

		// V2 endpoints wrap errors in the ApiResponse envelope. V1 Admin endpoints
		// return a bare JSON string, which fails to unmarshal here and leaves the
		// envelope fields empty — the status code and body still identify it.
		var errBody struct {
			Error *struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
			Metadata *struct {
				CorrelationID string `json:"correlationId"`
			} `json:"metadata"`
		}
		if jsonErr := json.Unmarshal(body, &errBody); jsonErr == nil && errBody.Error != nil {
			apiErr.Code = errBody.Error.Code
			apiErr.Message = errBody.Error.Message
			if errBody.Metadata != nil {
				apiErr.CorrelationID = errBody.Metadata.CorrelationID
			}
		}

		return nil, apiErr
	}

	return body, nil
}
