package azurenamingtool

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// serveV2 returns a server answering any path with the given data wrapped in the
// V2 envelope, recording which paths were requested.
func serveV2(t *testing.T, data interface{}, seen *[]string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if seen != nil {
			*seen = append(*seen, r.URL.Path)
		}
		body, err := json.Marshal(map[string]interface{}{"success": true, "data": data})
		if err != nil {
			t.Errorf("could not encode fixture: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
}

func TestGetResourceComponents(t *testing.T) {
	fixture := []map[string]interface{}{
		{"id": 1, "name": "ResourceType", "enabled": true, "sortOrder": 1, "isFreeText": false},
		{"id": 2, "name": "ResourceInstance", "enabled": true, "sortOrder": 9,
			"isFreeText": true, "minLength": "3", "maxLength": "3"},
	}
	var seen []string
	srv := serveV2(t, fixture, &seen)
	defer srv.Close()

	got, err := newTestClient(t, srv).GetResourceComponents()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d components, want 2", len(got))
	}
	if got[1].Name != "ResourceInstance" || !got[1].IsFreeText {
		t.Errorf("second component = %+v, want the free-text ResourceInstance", got[1])
	}
	if want := "/api/v2.0/ResourceComponents"; seen[0] != want {
		t.Errorf("requested %q, want %q", seen[0], want)
	}
}

// TestLengthRange covers the API carrying numeric bounds as strings, and leaving
// them empty when unconstrained -- which must not read as a zero-length limit.
func TestLengthRange(t *testing.T) {
	tests := map[string]struct {
		min, max         string
		wantMin, wantMax int
		wantOK           bool
	}{
		"fixed width": {"3", "3", 3, 3, true},
		"range":       {"1", "10", 1, 10, true},
		"padded":      {" 2 ", " 8 ", 2, 8, true},
		"both empty":  {"", "", 0, 0, false},
		"one empty":   {"3", "", 0, 0, false},
		"not numeric": {"three", "3", 0, 0, false},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := ResourceComponent{MinLength: tt.min, MaxLength: tt.max}
			min, max, ok := c.LengthRange()
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if ok && (min != tt.wantMin || max != tt.wantMax) {
				t.Errorf("range = (%d,%d), want (%d,%d)", min, max, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestEnabledComponentsOrders pins that components come back in the order they
// appear in a name, and that disabled ones are dropped. Both matter for building
// or checking a name.
func TestEnabledComponentsOrders(t *testing.T) {
	cfg := &NamingConfiguration{Components: []ResourceComponent{
		{Name: "ResourceInstance", Enabled: true, SortOrder: 9},
		{Name: "ResourceType", Enabled: true, SortOrder: 1},
		{Name: "Disabled", Enabled: false, SortOrder: 2},
		{Name: "ResourceOrg", Enabled: true, SortOrder: 5},
	}}

	got := []string{}
	for _, c := range cfg.EnabledComponents() {
		got = append(got, c.Name)
	}
	want := []string{"ResourceType", "ResourceOrg", "ResourceInstance"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("EnabledComponents() = %v, want %v", got, want)
	}
}

func TestDelimiterReturnsTheEnabledOne(t *testing.T) {
	cfg := &NamingConfiguration{Delimiters: []ResourceDelimiter{
		{Name: "underscore", Delimiter: "_", Enabled: false},
		{Name: "hyphen", Delimiter: "-", Enabled: true},
	}}
	if got := cfg.Delimiter(); got != "-" {
		t.Errorf("Delimiter() = %q, want %q", got, "-")
	}

	none := &NamingConfiguration{Delimiters: []ResourceDelimiter{
		{Name: "hyphen", Delimiter: "-", Enabled: false},
	}}
	if got := none.Delimiter(); got != "" {
		t.Errorf("with none enabled Delimiter() = %q, want empty", got)
	}
}

func TestComponentLookupIsCaseInsensitive(t *testing.T) {
	cfg := &NamingConfiguration{Components: []ResourceComponent{
		{Name: "ResourceInstance", MinLength: "3", MaxLength: "3"},
	}}
	for _, probe := range []string{"ResourceInstance", "resourceinstance", "RESOURCEINSTANCE"} {
		if _, ok := cfg.Component(probe); !ok {
			t.Errorf("Component(%q) not found", probe)
		}
	}
	if _, ok := cfg.Component("NoSuchComponent"); ok {
		t.Error("unknown component should not be found")
	}
}

func TestCustomComponentValues(t *testing.T) {
	cfg := &NamingConfiguration{CustomComponents: []CustomComponent{
		{ParentComponent: "subnet_tier", ShortName: "app"},
		{ParentComponent: "subnet_tier", ShortName: "web"},
		{ParentComponent: "application", ShortName: "webapp"},
		{ParentComponent: "subnet_tier", ShortName: ""}, // skipped
	}}
	if got, want := cfg.CustomComponentValues("subnet_tier"), []string{"app", "web"}; !reflect.DeepEqual(got, want) {
		t.Errorf("subnet_tier values = %v, want %v", got, want)
	}
	// Parent names come from configuration, so matching is case insensitive.
	if got := cfg.CustomComponentValues("SUBNET_TIER"); len(got) != 2 {
		t.Errorf("case-insensitive lookup returned %v", got)
	}
	if got := cfg.CustomComponentValues("nope"); len(got) != 0 {
		t.Errorf("unknown parent returned %v, want empty", got)
	}
}

func TestShortNames(t *testing.T) {
	got := ShortNames([]ComponentValue{
		{Name: "Production", ShortName: "p"},
		{Name: "Broken", ShortName: ""},
		{Name: "Development", ShortName: "d"},
	})
	if want := []string{"p", "d"}; !reflect.DeepEqual(got, want) {
		t.Errorf("ShortNames() = %v, want %v", got, want)
	}
}

// TestGetNamingConfigurationNamesTheFailingEndpoint keeps a partial outage
// diagnosable: without this the caller sees only that something was not found.
func TestGetNamingConfigurationNamesTheFailingEndpoint(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2.0/ResourceLocations" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(`{"error":{"code":"BOOM","message":"unavailable"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success":true,"data":[]}`))
	}))
	defer srv.Close()

	_, err := newTestClient(t, srv).GetNamingConfiguration()
	if err == nil {
		t.Fatal("expected an error")
	}
	if want := "ResourceLocations"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q should name the failing endpoint %q", err, want)
	}
}
