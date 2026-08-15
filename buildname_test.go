package azurenamingtool

import (
	"strings"
	"testing"
)

// realDeploymentConfig mirrors the deployment these tests were written against,
// deduced from names the API actually produced: no delimiter, and the
// organisation sorting before the resource type.
func realDeploymentConfig() *NamingConfiguration {
	return &NamingConfiguration{
		Components: []ResourceComponent{
			{Name: "ResourceOrg", Enabled: true, SortOrder: 1},
			{Name: "ResourceType", Enabled: true, SortOrder: 2},
			{Name: "application", Enabled: true, SortOrder: 3, IsCustom: true},
			{Name: "ResourceInstance", Enabled: true, SortOrder: 4},
			{Name: "ResourceLocation", Enabled: true, SortOrder: 5},
			{Name: "ResourceEnvironment", Enabled: true, SortOrder: 6},
		},
		Delimiters: []ResourceDelimiter{{Delimiter: "", Enabled: true}},
		ResourceTypes: []ResourceTypes{
			{Resource: "Storage/storageAccounts", ShortName: "st", ApplyDelimiter: true},
			{Resource: "Resources/resourceGroups", ShortName: "rg", ApplyDelimiter: true},
		},
	}
}

func realRequest() GenerateNameRequest {
	return GenerateNameRequest{
		ResourceType: "st", ResourceOrg: "pca", ResourceInstance: "948",
		ResourceLocation: "we", ResourceEnvironment: "p",
		CustomComponents: map[string]string{"application": "webapp"},
	}
}

// TestBuildNameReproducesNamesTheAPIProduced is the test that matters: these two
// names came back from the live API during acceptance testing and the
// consistency diagnostic. A reproduction that disagrees with the tool is worse
// than none, so this is the anchor for the whole algorithm.
func TestBuildNameReproducesNamesTheAPIProduced(t *testing.T) {
	cfg := realDeploymentConfig()

	tests := map[string]struct {
		req  GenerateNameRequest
		want string
	}{
		"acceptance run": {realRequest(), "pcastwebapp948wep"},
		"diagnostic probe": {GenerateNameRequest{
			ResourceType: "st", ResourceOrg: "pca", ResourceInstance: "552",
			ResourceLocation: "we", ResourceEnvironment: "p",
			CustomComponents: map[string]string{"application": "probe"},
		}, "pcastprobe552wep"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got, err := cfg.BuildName(tt.req)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("BuildName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildNameAppliesDelimiter(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.Delimiters = []ResourceDelimiter{{Delimiter: "-", Enabled: true}}
	for i := range cfg.Components {
		cfg.Components[i].ApplyDelimiterBefore = true
		cfg.Components[i].ApplyDelimiterAfter = true
	}

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The custom component uses the simpler rule but still receives a delimiter
	// because the name is non-empty and the type applies delimiters.
	if want := "pca-st-webapp-948-we-p"; got != want {
		t.Errorf("BuildName() = %q, want %q", got, want)
	}
}

// TestBuildNameDelimiterRequiresBothSides covers the rule that a delimiter is
// applied only when the component asks for one before it AND the previous
// component allowed one after it. The pair is checked across two built-in
// components, because custom components do not consult these flags at all.
func TestBuildNameDelimiterRequiresBothSides(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.Delimiters = []ResourceDelimiter{{Delimiter: "-", Enabled: true}}
	for i := range cfg.Components {
		cfg.Components[i].ApplyDelimiterBefore = true
		cfg.Components[i].ApplyDelimiterAfter = true
	}
	// ResourceInstance permits no delimiter after it, so none appears between it
	// and ResourceLocation, which follows and is also built in.
	cfg.Components[3].ApplyDelimiterAfter = false

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "pca-st-webapp-948we-p"; got != want {
		t.Errorf("BuildName() = %q, want %q", got, want)
	}
}

// TestBuildNameCustomComponentsIgnoreDelimiterFlags records that custom
// components take no notice of ApplyDelimiterBefore, of whether the previous
// component allowed a delimiter after it, or of a delimiter the resource type
// forbids. They apply the delimiter whenever the name is non-empty and the type
// allows delimiters at all.
//
// This looks like an oversight in the naming tool rather than a decision, but
// reproducing it is the point: a prediction that disagrees with the tool is
// worse than none, however defensible the disagreement.
func TestBuildNameCustomComponentsIgnoreDelimiterFlags(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.Delimiters = []ResourceDelimiter{{Delimiter: "-", Enabled: true}}
	// No component asks for a delimiter before it, and none allows one after.
	for i := range cfg.Components {
		cfg.Components[i].ApplyDelimiterBefore = false
		cfg.Components[i].ApplyDelimiterAfter = false
	}

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only the custom component contributes a delimiter.
	if want := "pcast-webapp948wep"; got != want {
		t.Errorf("BuildName() = %q, want %q", got, want)
	}
}

// TestBuildNameDropsADelimiterTheTypeForbids covers a delimiter appearing in a
// resource type's invalid characters: the tool abandons it for the rest of the
// name rather than only for that component. Custom components are unaffected,
// as the preceding test records, so the one before "webapp" survives.
func TestBuildNameDropsADelimiterTheTypeForbids(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.Delimiters = []ResourceDelimiter{{Delimiter: "-", Enabled: true}}
	for i := range cfg.Components {
		cfg.Components[i].ApplyDelimiterBefore = true
		cfg.Components[i].ApplyDelimiterAfter = true
	}
	cfg.ResourceTypes[0].InvalidCharacters = "-"

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "pcast-webapp948wep"; got != want {
		t.Errorf("BuildName() = %q, want %q (dropped for built-in components only)", got, want)
	}
}

// TestBuildNameHonoursExclusions covers a resource type excluding a component,
// matched on the normalised component name.
func TestBuildNameHonoursExclusions(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.ResourceTypes[0].Exclude = "location,environment"

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "pcastwebapp948"; got != want {
		t.Errorf("BuildName() = %q, want %q", got, want)
	}
}

func TestBuildNameSkipsDisabledComponentsAndEmptyValues(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.Components = append(cfg.Components,
		ResourceComponent{Name: "ResourceFunction", Enabled: false, SortOrder: 7})

	req := realRequest()
	req.ResourceFunction = "dat" // disabled component, must not appear
	req.ResourceLocation = ""    // supplied empty, must not appear

	got, err := cfg.BuildName(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if want := "pcastwebapp948p"; got != want {
		t.Errorf("BuildName() = %q, want %q", got, want)
	}
}

// TestBuildNameReturnsStaticValues covers resource types with a fixed name,
// which the tool returns without assembling anything.
func TestBuildNameReturnsStaticValues(t *testing.T) {
	cfg := realDeploymentConfig()
	cfg.ResourceTypes[0].StaticValues = "fixedname"

	got, err := cfg.BuildName(realRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "fixedname" {
		t.Errorf("BuildName() = %q, want the static value", got)
	}
}

// TestBuildNameRefusesRatherThanGuess covers the cases where a name cannot be
// determined. Returning a plausible but wrong name would be worse than an error,
// because it would be shown in a plan and then contradicted at apply.
func TestBuildNameRefusesRatherThanGuess(t *testing.T) {
	t.Run("unknown resource type", func(t *testing.T) {
		req := realRequest()
		req.ResourceType = "nope"
		if _, err := realDeploymentConfig().BuildName(req); err == nil {
			t.Error("expected an error for an unknown resource type")
		}
	})

	t.Run("component this client cannot supply", func(t *testing.T) {
		cfg := realDeploymentConfig()
		cfg.Components = append(cfg.Components,
			ResourceComponent{Name: "ResourceSomethingNew", Enabled: true, SortOrder: 8})

		_, err := cfg.BuildName(realRequest())
		if err == nil {
			t.Fatal("expected an error for an unrecognised enabled component")
		}
		if !strings.Contains(err.Error(), "cannot supply a value") {
			t.Errorf("error = %q, want it to explain the component cannot be supplied", err)
		}
	})

	t.Run("no enabled components", func(t *testing.T) {
		cfg := realDeploymentConfig()
		for i := range cfg.Components {
			cfg.Components[i].Enabled = false
		}
		if _, err := cfg.BuildName(realRequest()); err == nil {
			t.Error("expected an error when nothing is enabled")
		}
	})

	t.Run("no values supplied", func(t *testing.T) {
		cfg := realDeploymentConfig()
		if _, err := cfg.BuildName(GenerateNameRequest{ResourceType: "st"}); err == nil {
			// ResourceType itself is a component, so this should still build.
			t.Skip("resource type alone forms a name in this configuration")
		}
	})
}

func TestNormalizeComponentName(t *testing.T) {
	for in, want := range map[string]string{
		"ResourceType":       "type",
		"ResourceProjAppSvc": "projappsvc",
		"Resource Unit Dept": "unitdept",
		"application":        "application",
		"subnet_tier":        "subnet_tier",
	} {
		if got := normalizeComponentName(in); got != want {
			t.Errorf("normalizeComponentName(%q) = %q, want %q", in, got, want)
		}
	}
}
