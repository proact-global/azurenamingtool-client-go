package azurenamingtool

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func strptr(s string) *string { return &s }

// TestNewClientRejectsMalformedHosts covers the values that would otherwise fail
// only once a request is attempted, with an error that points nowhere useful.
func TestNewClientRejectsMalformedHosts(t *testing.T) {
	tests := map[string]struct {
		host        string
		wantMessage string
	}{
		// The exact shape of a host configured without its protocol. Go reports
		// this as `unsupported protocol scheme ""` from inside the first request.
		"missing scheme": {
			host:        "naming-tool.example.com",
			wantMessage: "missing a scheme",
		},
		"missing scheme with path": {
			host:        "naming-tool.example.com/api",
			wantMessage: "missing a scheme",
		},
		"empty": {
			host:        "",
			wantMessage: "host is empty",
		},
		"whitespace only": {
			host:        "   ",
			wantMessage: "host is empty",
		},
		"unsupported scheme": {
			host:        "ftp://naming-tool.example.com",
			wantMessage: "unsupported scheme",
		},
		"scheme but no host": {
			host:        "https://",
			wantMessage: "does not contain a host name",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := NewClient(strptr(tt.host), strptr("key"), nil)
			if err == nil {
				t.Fatalf("expected an error for host %q, got client with HostURL %q", tt.host, c.HostURL)
			}
			if !strings.Contains(err.Error(), tt.wantMessage) {
				t.Errorf("error = %q, want it to mention %q", err.Error(), tt.wantMessage)
			}
		})
	}
}

func TestNewClientNormalizesHost(t *testing.T) {
	tests := map[string]struct{ host, want string }{
		"plain https":            {"https://naming.example.com", "https://naming.example.com"},
		"trailing slash removed": {"https://naming.example.com/", "https://naming.example.com"},
		"several slashes":        {"https://naming.example.com///", "https://naming.example.com"},
		"surrounding whitespace": {"  https://naming.example.com  ", "https://naming.example.com"},
		"http is allowed":        {"http://localhost:19090", "http://localhost:19090"},
		"sub path preserved":     {"https://naming.example.com/tool", "https://naming.example.com/tool"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c, err := NewClient(strptr(tt.host), strptr("key"), nil)
			if err != nil {
				t.Fatalf("unexpected error for host %q: %v", tt.host, err)
			}
			if c.HostURL != tt.want {
				t.Errorf("HostURL = %q, want %q", c.HostURL, tt.want)
			}
		})
	}
}

// TestNewClientDefaultsWhenHostIsNil keeps nil host working: it falls back to
// the package default, which must itself pass validation.
func TestNewClientDefaultsWhenHostIsNil(t *testing.T) {
	c, err := NewClient(nil, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.HostURL != HostURL {
		t.Errorf("HostURL = %q, want the default %q", c.HostURL, HostURL)
	}
	if c.APIKey != "" || c.AdminPassword != nil {
		t.Errorf("nil credentials should stay unset, got APIKey=%q AdminPassword=%v", c.APIKey, c.AdminPassword)
	}
}

// TestTrailingSlashDoesNotDoubleUpInRequestPaths is the reason normalisation
// happens at construction: request URLs are built by concatenation, so a
// trailing slash would otherwise appear in every path.
func TestTrailingSlashDoesNotDoubleUpInRequestPaths(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"id":1}`))
	}))
	defer srv.Close()

	c, err := NewClient(strptr(srv.URL+"/"), strptr("key"), strptr("admin"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := c.GetName(1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "/api/Admin/GetGeneratedName/1"; gotPath != want {
		t.Errorf("request path = %q, want %q", gotPath, want)
	}
}
