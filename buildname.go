package azurenamingtool

import (
	"fmt"
	"strings"
)

// BuildName works out the name the naming tool would generate for a request,
// without asking it to generate one.
//
// The naming tool has no endpoint that returns a name without also storing it,
// so the only way to show a name before committing to it is to reproduce how one
// is assembled. This does that, following the tool's own algorithm and driven
// entirely by the configuration read from it, so a deployment's components,
// ordering, delimiter and exclusions are all respected.
//
// It is a reproduction, and that carries an obligation: a result that disagrees
// with the tool is worse than no result at all. Where the algorithm depends on
// something this cannot see, BuildName returns an error rather than a guess.
//
// The tool's assembly is in ResourceNamingRequestService.RequestNameAsync. The
// behaviours reproduced here that are easy to overlook:
//
//   - A resource type carrying StaticValues has a fixed name; nothing is assembled.
//   - Components are those enabled, in SortOrder.
//   - A resource type may exclude components by normalised name.
//   - The delimiter is applied before a component only when the component asks
//     for it, the previous component allowed one after it, the resource type
//     applies delimiters at all, and the name is not empty.
//   - A delimiter that appears in a resource type's invalid characters is
//     dropped for the remainder of the name, not just that component.
//   - Custom components use simpler delimiter handling: the delimiter is applied
//     whenever the name is non-empty and the resource type allows delimiters,
//     regardless of the component's own flags.
//   - The flag recording whether the previous component allowed a delimiter
//     after it is updated for every component, including excluded ones and those
//     with no value.
func (n *NamingConfiguration) BuildName(req GenerateNameRequest) (string, error) {
	resourceType, ok := n.resourceTypeByShortName(req.ResourceType)
	if !ok {
		return "", fmt.Errorf("no resource type has the short name %q", req.ResourceType)
	}

	// A static resource type has a fixed name and is not assembled at all.
	if strings.TrimSpace(resourceType.StaticValues) != "" {
		return resourceType.StaticValues, nil
	}

	components := n.EnabledComponents()
	if len(components) == 0 {
		return "", fmt.Errorf("the naming tool has no enabled components, so no name can be built")
	}

	delimiter := n.Delimiter()
	excluded := splitLowerCSV(resourceType.Exclude)

	var name strings.Builder
	ignoreDelimiter := false
	// The tool starts as though the previous component permitted a delimiter
	// after it, so the first delimiter decision rests on the component itself.
	previousAllowedDelimiterAfter := true

	for _, component := range components {
		normalized := normalizeComponentName(component.Name)

		// Excluded components contribute nothing, but still affect the delimiter
		// state for the component that follows.
		if excluded[normalized] {
			previousAllowedDelimiterAfter = component.ApplyDelimiterAfter
			continue
		}

		value, err := componentValue(component, normalized, req)
		if err != nil {
			return "", err
		}

		if value == "" {
			previousAllowedDelimiterAfter = component.ApplyDelimiterAfter
			continue
		}

		if component.IsCustom {
			// Custom components ignore the per-component delimiter flags.
			if name.Len() > 0 && resourceType.ApplyDelimiter {
				name.WriteString(delimiter)
			}
		} else {
			applyDelimiter := !ignoreDelimiter &&
				delimiter != "" &&
				name.Len() > 0 &&
				component.ApplyDelimiterBefore &&
				previousAllowedDelimiterAfter &&
				resourceType.ApplyDelimiter

			// A delimiter the resource type forbids is abandoned for the rest of
			// the name, matching the tool, which sets its ignore flag and never
			// clears it.
			if !ignoreDelimiter && delimiter != "" &&
				resourceType.InvalidCharacters != "" &&
				strings.Contains(resourceType.InvalidCharacters, delimiter) {
				ignoreDelimiter = true
				applyDelimiter = false
			}

			if applyDelimiter {
				name.WriteString(delimiter)
			}
		}

		name.WriteString(value)
		previousAllowedDelimiterAfter = component.ApplyDelimiterAfter
	}

	if name.Len() == 0 {
		return "", fmt.Errorf("the request supplied no values for any enabled component")
	}

	return name.String(), nil
}

// componentValue returns the request's value for a component, by the same route
// the naming tool takes: built-in components read a named field, custom ones a
// map entry keyed by the normalised component name.
func componentValue(component ResourceComponent, normalized string, req GenerateNameRequest) (string, error) {
	if component.IsCustom {
		return req.CustomComponents[normalized], nil
	}

	switch normalized {
	case "type":
		return req.ResourceType, nil
	case "org":
		return req.ResourceOrg, nil
	case "environment":
		return req.ResourceEnvironment, nil
	case "location":
		return req.ResourceLocation, nil
	case "function":
		return req.ResourceFunction, nil
	case "instance":
		return req.ResourceInstance, nil
	case "projappsvc":
		return req.ResourceProjAppSvc, nil
	case "unitdept":
		return req.ResourceUnitDept, nil
	}

	// A built-in component this client does not know how to supply. Guessing
	// would produce a name that silently omits it, so refuse instead: a wrong
	// name is worse than none.
	return "", fmt.Errorf(
		"component %q is enabled but this client cannot supply a value for it, "+
			"so the generated name cannot be predicted", component.Name)
}

// normalizeComponentName mirrors the tool's own normalisation, which strips the
// "Resource" prefix and spaces and lowercases the result. It is what a resource
// type's exclusions and a request's custom component keys are matched on.
func normalizeComponentName(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(name, "Resource", ""), " ", ""))
}

// splitLowerCSV turns a comma-separated list into a set, lowercased. Used for a
// resource type's excluded and optional component lists.
func splitLowerCSV(s string) map[string]bool {
	out := map[string]bool{}
	for _, part := range strings.Split(strings.ToLower(s), ",") {
		if p := strings.TrimSpace(part); p != "" {
			out[p] = true
		}
	}
	return out
}

// resourceTypeByShortName finds a resource type by its short name, matched
// exactly, as the naming tool does.
func (n *NamingConfiguration) resourceTypeByShortName(shortName string) (ResourceTypes, bool) {
	for _, rt := range n.ResourceTypes {
		if rt.ShortName == shortName {
			return rt, true
		}
	}
	return ResourceTypes{}, false
}
