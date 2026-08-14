package azurenamingtool

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient returns a Client pointed at the given test server.
func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()

	host := srv.URL
	apiKey := "test-api-key"
	adminPwd := "test-admin-password"

	c, err := NewClient(&host, &apiKey, &adminPwd)
	if err != nil {
		t.Fatalf("NewClient returned an unexpected error: %v", err)
	}
	return c
}

// respondWith returns a handler that writes a fixed status and body.
func respondWith(status int, body string) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestGetNameMapsAdminNotFoundToErrNotFound(t *testing.T) {
	// The V1 Admin endpoint reports a missing ID as HTTP 400 with a bare JSON
	// string body. The message is pluralised when the store itself is empty.
	tests := map[string]string{
		"singular": `"Generated Name not found!"`,
		"plural":   `"Generated Names not found!"`,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(respondWith(http.StatusBadRequest, body))
			defer srv.Close()

			_, err := newTestClient(t, srv).GetName(42)
			if !errors.Is(err, ErrNotFound) {
				t.Fatalf("expected ErrNotFound, got %v", err)
			}
		})
	}
}

// TestGetNameDoesNotMapUnrelatedFailuresToErrNotFound is the important one.
// Callers treat ErrNotFound as "this name is gone, regenerate it", so mistaking
// an infrastructure error for a missing record silently replaces a name that is
// still in use. Only the naming tool's own message may produce ErrNotFound.
func TestGetNameDoesNotMapUnrelatedFailuresToErrNotFound(t *testing.T) {
	tests := map[string]struct {
		status int
		body   string
	}{
		"proxy 404 page": {
			http.StatusNotFound,
			"<html><body>404 - File or directory not found.</body></html>",
		},
		"moved route": {
			http.StatusNotFound,
			`{"error":{"code":"NOT_FOUND","message":"The requested endpoint was not found"}}`,
		},
		"auth failure": {
			http.StatusUnauthorized,
			"You do not have permission to view this directory or page.",
		},
		"server error": {
			http.StatusInternalServerError,
			`{"error":{"code":"INTERNAL_SERVER_ERROR","message":"boom"}}`,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(respondWith(tc.status, tc.body))
			defer srv.Close()

			_, err := newTestClient(t, srv).GetName(42)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
			if errors.Is(err, ErrNotFound) {
				t.Fatalf("unrelated failure was misreported as ErrNotFound: %v", err)
			}
		})
	}
}

func TestGetNameSuccess(t *testing.T) {
	const body = `{"id":123,"resourceName":"stmanwebappeuwdev001",` +
		`"resourceTypeName":"Storage account","components":[["ResourceType","st"]]}`

	srv := httptest.NewServer(respondWith(http.StatusOK, body))
	defer srv.Close()

	got, err := newTestClient(t, srv).GetName(123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != 123 {
		t.Errorf("ID = %d, want 123", got.ID)
	}
	if got.ResourceName != "stmanwebappeuwdev001" {
		t.Errorf("ResourceName = %q, want %q", got.ResourceName, "stmanwebappeuwdev001")
	}
	if got.ResourceTypeName != "Storage account" {
		t.Errorf("ResourceTypeName = %q, want %q", got.ResourceTypeName, "Storage account")
	}
}

func TestDeleteNameTreatsNotFoundAsSuccess(t *testing.T) {
	// Deleting an entry that is already gone satisfies the caller's intent, so it
	// must not surface as an error. Both message forms are accepted.
	for name, body := range map[string]string{
		"singular": `"Generated Name not found!"`,
		"plural":   `"Generated Names not found!"`,
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(respondWith(http.StatusBadRequest, body))
			defer srv.Close()

			if _, err := newTestClient(t, srv).DeleteName(DeleteGeneratedNameRequest{ID: 42}); err != nil {
				t.Fatalf("expected nil error for an already-deleted entry, got %v", err)
			}
		})
	}
}

// TestDeleteNameReportsAdminPasswordFailure covers the naming tool's habit of
// answering HTTP 200 with a "FAILURE" body when the admin password is wrong.
// Callers rely on this surfacing as an error.
func TestDeleteNameReportsAdminPasswordFailure(t *testing.T) {
	for name, body := range map[string]string{
		"incorrect password": "FAILURE - Incorrect Global Admin Password.",
		"missing password":   "FAILURE - You must provide the Global Admin Password.",
	} {
		t.Run(name, func(t *testing.T) {
			srv := httptest.NewServer(respondWith(http.StatusOK, body))
			defer srv.Close()

			if _, err := newTestClient(t, srv).DeleteName(DeleteGeneratedNameRequest{ID: 42}); err == nil {
				t.Fatal("expected an error for a FAILURE response, got nil")
			}
		})
	}
}

func TestDeleteNamePropagatesUnrelatedFailures(t *testing.T) {
	srv := httptest.NewServer(respondWith(http.StatusInternalServerError, "boom"))
	defer srv.Close()

	if _, err := newTestClient(t, srv).DeleteName(DeleteGeneratedNameRequest{ID: 42}); err == nil {
		t.Fatal("expected a server error to propagate, got nil")
	}
}

func TestAPIErrorCarriesStatusAndEnvelope(t *testing.T) {
	const body = `{"error":{"code":"NAME_GENERATION_FAILED","message":"bad input"},` +
		`"metadata":{"correlationId":"abc-123"}}`

	srv := httptest.NewServer(respondWith(http.StatusBadRequest, body))
	defer srv.Close()

	_, err := newTestClient(t, srv).GenerateName(GenerateNameRequest{})

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected an *APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusBadRequest)
	}
	if apiErr.Code != "NAME_GENERATION_FAILED" {
		t.Errorf("Code = %q, want %q", apiErr.Code, "NAME_GENERATION_FAILED")
	}
	if apiErr.CorrelationID != "abc-123" {
		t.Errorf("CorrelationID = %q, want %q", apiErr.CorrelationID, "abc-123")
	}
}

// TestAdminPasswordHeaderIsSent guards the header name the Admin API binds to.
func TestAdminPasswordHeaderIsSent(t *testing.T) {
	var gotAdminPassword, gotAPIKey string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAdminPassword = r.Header.Get("AdminPassword")
		gotAPIKey = r.Header.Get("APIKey")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	if _, err := newTestClient(t, srv).GetName(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotAdminPassword != "test-admin-password" {
		t.Errorf("AdminPassword header = %q, want %q", gotAdminPassword, "test-admin-password")
	}
	if gotAPIKey != "test-api-key" {
		t.Errorf("APIKey header = %q, want %q", gotAPIKey, "test-api-key")
	}
}
