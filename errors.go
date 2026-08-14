package azurenamingtool

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

// ErrNotFound is returned by GetName when the requested ID does not exist in the naming tool.
var ErrNotFound = errors.New("generated name not found")

// APIError describes a non-200 response from the Azure Naming Tool API. It keeps
// the HTTP status code and the raw body alongside the parsed V2 error envelope so
// callers can branch on the status code instead of matching on error text.
type APIError struct {
	// StatusCode is the HTTP status returned by the API.
	StatusCode int
	// Code and Message are populated from the V2 ApiResponse error envelope when
	// the response carries one. V1 Admin endpoints return bare strings and leave
	// these empty.
	Code    string
	Message string
	// CorrelationID comes from the V2 response metadata, when present.
	CorrelationID string
	// Body is the raw response body.
	Body string
}

func (e *APIError) Error() string {
	if e.Code != "" {
		msg := fmt.Sprintf("[%s] %s", e.Code, e.Message)
		if e.CorrelationID != "" {
			msg += fmt.Sprintf(" (correlationId: %s)", e.CorrelationID)
		}
		return msg
	}
	return fmt.Sprintf("status: %d, body: %s", e.StatusCode, e.Body)
}

// generatedNameNotFoundRe matches the message the Admin API returns for a missing
// generated name. The message is pluralised when the store itself is empty
// ("Generated Names not found!"), so both forms are accepted.
var generatedNameNotFoundRe = regexp.MustCompile(`(?i)generated names? not found`)

// isNotFound reports whether err is the Admin API's "this generated name does not
// exist" response.
//
// It deliberately requires both a not-found-ish status code and the naming tool's
// own message. Matching on the message alone would misread any unrelated 404 — a
// moved Admin route, a mistyped host, an intermediate proxy's error page — as a
// missing record. Callers treat ErrNotFound as "this name is gone, regenerate it",
// so a false positive there silently replaces a name that is still in use.
func isNotFound(err error) bool {
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		return false
	}
	// The Admin endpoint reports a missing ID as HTTP 400; 404 is accepted too in
	// case a future version uses the more conventional status.
	if apiErr.StatusCode != http.StatusBadRequest && apiErr.StatusCode != http.StatusNotFound {
		return false
	}
	return generatedNameNotFoundRe.MatchString(apiErr.Body)
}
