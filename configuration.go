package azurenamingtool

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// This file exposes the naming tool's own configuration: which components make
// up a name, in what order, with which delimiter, and which values each
// component accepts.
//
// It is what allows a caller to check a request before sending it, rather than
// discovering from a failed generation that a short name does not exist or that
// a component is the wrong length. Every endpoint here is a read, authenticated
// by the API key alone -- no admin password is involved.

// ResourceComponent describes one component of a generated name.
type ResourceComponent struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Description string `json:"description"`

	// Enabled reports whether this component takes part in a name at all.
	Enabled bool `json:"enabled"`
	// SortOrder is the position of the component within the name.
	SortOrder int `json:"sortOrder"`

	// IsCustom marks a component defined by the deployment rather than built in.
	IsCustom bool `json:"isCustom"`
	// IsFreeText marks a component accepting any value. When false the accepted
	// values are enumerable, so a value can be checked exactly.
	IsFreeText bool `json:"isFreeText"`

	// MinLength and MaxLength are numeric but carried as strings by the API.
	// Use LengthRange to read them.
	MinLength string `json:"minLength"`
	MaxLength string `json:"maxLength"`

	EnforceRandom bool `json:"enforceRandom"`
	Alphanumeric  bool `json:"alphanumeric"`

	// ApplyDelimiterBefore and ApplyDelimiterAfter govern where the delimiter is
	// placed relative to this component.
	ApplyDelimiterBefore bool `json:"applyDelimiterBefore"`
	ApplyDelimiterAfter  bool `json:"applyDelimiterAfter"`
}

// LengthRange returns the component's permitted length as numbers. ok is false
// when the deployment leaves the bounds unset, which means unconstrained rather
// than zero-length.
func (c ResourceComponent) LengthRange() (minLen, maxLen int, ok bool) {
	minStr, maxStr := strings.TrimSpace(c.MinLength), strings.TrimSpace(c.MaxLength)
	if minStr == "" || maxStr == "" {
		return 0, 0, false
	}
	minVal, errMin := strconv.Atoi(minStr)
	maxVal, errMax := strconv.Atoi(maxStr)
	if errMin != nil || errMax != nil {
		return 0, 0, false
	}
	return minVal, maxVal, true
}

// ResourceDelimiter is the separator placed between components.
type ResourceDelimiter struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Delimiter string `json:"delimiter"`
	Enabled   bool   `json:"enabled"`
	SortOrder int    `json:"sortOrder"`
}

// CustomComponent is one accepted value for a custom component, tied to its
// parent component by ParentComponent.
type CustomComponent struct {
	ID              int    `json:"id"`
	ParentComponent string `json:"parentComponent"`
	Name            string `json:"name"`
	ShortName       string `json:"shortName"`
	SortOrder       int    `json:"sortOrder"`
	MinLength       string `json:"minLength"`
	MaxLength       string `json:"maxLength"`
}

// ComponentValue is one accepted value for a built-in component. Organisations,
// locations, environments, functions, project/app services and unit/departments
// all share this shape; Enabled is only meaningful for locations, where the API
// populates it.
type ComponentValue struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"shortName"`
	SortOrder int    `json:"sortOrder"`
	Enabled   bool   `json:"enabled"`
}

// getV2 performs a GET against a V2 endpoint and unwraps the ApiResponse
// envelope. It is a package-level function rather than a method because Go does
// not allow type parameters on methods.
func getV2[T any](c *Client, path string) (T, error) {
	var zero T

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.HostURL, path), nil)
	if err != nil {
		return zero, err
	}

	c.mu.Lock()
	body, err := c.doRequest(req)
	c.mu.Unlock()
	if err != nil {
		return zero, err
	}

	resp := ApiResponse[T]{}
	if err := json.Unmarshal(body, &resp); err != nil {
		return zero, err
	}
	if !resp.Success {
		if resp.Error != nil {
			return zero, fmt.Errorf("[%s] %s", resp.Error.Code, resp.Error.Message)
		}
		return zero, fmt.Errorf("request to %s failed", path)
	}

	return resp.Data, nil
}

// GetResourceComponents returns the components a name is built from, including
// disabled ones. Callers that care only about components in use should filter on
// Enabled.
func (c *Client) GetResourceComponents() ([]ResourceComponent, error) {
	return getV2[[]ResourceComponent](c, "/api/v2.0/ResourceComponents")
}

// GetResourceDelimiters returns the configured delimiters. The one in use is the
// enabled entry.
func (c *Client) GetResourceDelimiters() ([]ResourceDelimiter, error) {
	return getV2[[]ResourceDelimiter](c, "/api/v2.0/ResourceDelimiters")
}

// GetCustomComponents returns every custom component value across all parents.
func (c *Client) GetCustomComponents() ([]CustomComponent, error) {
	return getV2[[]CustomComponent](c, "/api/v2.0/CustomComponents")
}

// GetResourceOrgs returns the accepted organisation values.
func (c *Client) GetResourceOrgs() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceOrgs")
}

// GetResourceLocations returns the accepted location values.
func (c *Client) GetResourceLocations() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceLocations")
}

// GetResourceEnvironments returns the accepted environment values.
func (c *Client) GetResourceEnvironments() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceEnvironments")
}

// GetResourceFunctions returns the accepted function values.
func (c *Client) GetResourceFunctions() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceFunctions")
}

// GetResourceProjAppSvcs returns the accepted project/app service values.
func (c *Client) GetResourceProjAppSvcs() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceProjAppSvcs")
}

// GetResourceUnitDepts returns the accepted unit/department values.
func (c *Client) GetResourceUnitDepts() ([]ComponentValue, error) {
	return getV2[[]ComponentValue](c, "/api/v2.0/ResourceUnitDepts")
}

// NamingConfiguration is everything needed to check a request, or to work out
// what a generated name will look like, in one value.
type NamingConfiguration struct {
	Components       []ResourceComponent
	Delimiters       []ResourceDelimiter
	ResourceTypes    []ResourceTypes
	CustomComponents []CustomComponent

	Orgs         []ComponentValue
	Locations    []ComponentValue
	Environments []ComponentValue
	Functions    []ComponentValue
	ProjAppSvcs  []ComponentValue
	UnitDepts    []ComponentValue
}

// GetNamingConfiguration fetches the whole configuration.
//
// The requests are sequential because the client serialises requests anyway, and
// because a naming tool is a small single-instance service that is better not
// hit with a burst. Callers are expected to fetch this once and reuse it, not
// call it per name.
func (c *Client) GetNamingConfiguration() (*NamingConfiguration, error) {
	cfg := &NamingConfiguration{}

	// Each step names the endpoint it failed on, since "not found" from a bare
	// GET says nothing about which part of the configuration is unavailable.
	steps := []struct {
		name string
		load func() error
	}{
		{"ResourceComponents", func() (err error) { cfg.Components, err = c.GetResourceComponents(); return }},
		{"ResourceDelimiters", func() (err error) { cfg.Delimiters, err = c.GetResourceDelimiters(); return }},
		{"ResourceTypes", func() (err error) { cfg.ResourceTypes, err = c.GetResourceTypes(); return }},
		{"CustomComponents", func() (err error) { cfg.CustomComponents, err = c.GetCustomComponents(); return }},
		{"ResourceOrgs", func() (err error) { cfg.Orgs, err = c.GetResourceOrgs(); return }},
		{"ResourceLocations", func() (err error) { cfg.Locations, err = c.GetResourceLocations(); return }},
		{"ResourceEnvironments", func() (err error) { cfg.Environments, err = c.GetResourceEnvironments(); return }},
		{"ResourceFunctions", func() (err error) { cfg.Functions, err = c.GetResourceFunctions(); return }},
		{"ResourceProjAppSvcs", func() (err error) { cfg.ProjAppSvcs, err = c.GetResourceProjAppSvcs(); return }},
		{"ResourceUnitDepts", func() (err error) { cfg.UnitDepts, err = c.GetResourceUnitDepts(); return }},
	}

	for _, s := range steps {
		if err := s.load(); err != nil {
			return nil, fmt.Errorf("reading %s: %w", s.name, err)
		}
	}

	return cfg, nil
}

// EnabledComponents returns the components that take part in a name, ordered as
// they appear in it.
func (n *NamingConfiguration) EnabledComponents() []ResourceComponent {
	out := make([]ResourceComponent, 0, len(n.Components))
	for _, c := range n.Components {
		if c.Enabled {
			out = append(out, c)
		}
	}
	// Insertion sort: the list is short and already close to ordered.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].SortOrder < out[j-1].SortOrder; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

// Delimiter returns the delimiter in use, or an empty string when none is
// enabled.
func (n *NamingConfiguration) Delimiter() string {
	for _, d := range n.Delimiters {
		if d.Enabled {
			return d.Delimiter
		}
	}
	return ""
}

// Component returns the component with the given name, matched case
// insensitively because component names are configuration rather than
// identifiers.
func (n *NamingConfiguration) Component(name string) (ResourceComponent, bool) {
	for _, c := range n.Components {
		if strings.EqualFold(c.Name, name) {
			return c, true
		}
	}
	return ResourceComponent{}, false
}

// ShortNames returns the accepted short names for a list of component values.
func ShortNames(values []ComponentValue) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v.ShortName != "" {
			out = append(out, v.ShortName)
		}
	}
	return out
}

// CustomComponentValues returns the accepted short names for a custom component,
// identified by its parent component name.
func (n *NamingConfiguration) CustomComponentValues(parent string) []string {
	out := []string{}
	for _, cc := range n.CustomComponents {
		if strings.EqualFold(cc.ParentComponent, parent) && cc.ShortName != "" {
			out = append(out, cc.ShortName)
		}
	}
	return out
}
