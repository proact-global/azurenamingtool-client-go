package azurenamingtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

// NewClient returns a client for the Azure Naming Tool at host. Pass nil for
// host to use the default, or for AdminPassword when admin operations (delete,
// lookup by ID) are not needed.
//
// The host is validated and normalised here so that a malformed value is
// reported at construction rather than as an opaque failure on the first
// request. See normalizeHostURL.
func NewClient(host, apiKey, AdminPassword *string) (*Client, error) {
	c := Client{
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		HostURL:    HostURL,
	}

	if host != nil {
		c.HostURL = *host
	}

	normalized, err := normalizeHostURL(c.HostURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Azure Naming Tool host: %w", err)
	}
	c.HostURL = normalized

	if apiKey != nil {
		c.APIKey = *apiKey
	}

	if AdminPassword != nil {
		c.AdminPassword = AdminPassword
	}

	return &c, nil
}

// normalizeHostURL checks that host is a usable base URL and returns it without
// trailing slashes.
//
// Request URLs are built by string concatenation elsewhere in this package, so
// both checks earn their keep at construction time. A host with no scheme
// otherwise surfaces only once a request is attempted, as Go's opaque
// `unsupported protocol scheme ""` raised from deep inside an unrelated call,
// and a trailing slash silently produces a doubled slash in every path.
func normalizeHostURL(host string) (string, error) {
	trimmed := strings.TrimSpace(host)
	if trimmed == "" {
		return "", errors.New("host is empty")
	}

	parsed, err := url.Parse(trimmed)
	if err != nil {
		return "", fmt.Errorf("host %q is not a valid URL: %w", trimmed, err)
	}

	switch parsed.Scheme {
	case "http", "https":
	case "":
		return "", fmt.Errorf(
			"host %q is missing a scheme, expected something like \"https://%s\"", trimmed, trimmed)
	default:
		return "", fmt.Errorf(
			"host %q has unsupported scheme %q, expected \"http\" or \"https\"", trimmed, parsed.Scheme)
	}

	if parsed.Host == "" {
		return "", fmt.Errorf("host %q does not contain a host name", trimmed)
	}

	return strings.TrimRight(trimmed, "/"), nil
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
