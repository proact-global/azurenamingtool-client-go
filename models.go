package azurenamingtool

// ResourceTypes -
type ResourceTypes struct {
	ID                           int    `json:"id"`
	Resource                     string `json:"resource"`
	Optional                     string `json:"optional"`
	Exclude                      string `json:"exclude"`
	Property                     string `json:"property"`
	ShortName                    string `json:"ShortName"`
	Scope                        string `json:"scope"`
	LengthMin                    string `json:"lengthMin"`
	LengthMax                    string `json:"lengthMax"`
	ValidText                    string `json:"validText"`
	InvalidText                  string `json:"invalidText"`
	InvalidCharacters            string `json:"invalidCharacters"`
	InvalidCharactersStart       string `json:"invalidCharactersStart"`
	InvalidCharactersEnd         string `json:"invalidCharactersEnd"`
	InvalidCharactersConsecutive string `json:"invalidCharactersConsecutive"`
	Regx                         string `json:"regx"`
	StaticValues                 string `json:"staticValues"`
	Enabled                      bool   `json:"enabled"`
	ApplyDelimiter               bool   `json:"applyDelimiter"`
}

type GenerateNameRequest struct {
	ResourceEnvironment string                              `json:"resourceEnvironment"`
	ResourceFunction    string                              `json:"resourceFunction"`
	ResourceInstance    string                              `json:"resourceInstance"`
	ResourceLocation    string                              `json:"resourceLocation"`
	ResourceOrg         string                              `json:"resourceOrg"`
	ResourceType        string                              `json:"resourceType"`
	CustomComponents    GenerateNameRequestCustomComponents `json:"customComponents"`
}

type GenerateNameRequestCustomComponents struct {
	Application string `json:"application"`
}

type GenerateNameResponse struct {
	ResourceName        string              `json:"resourceName"`
	Message             string              `json:"message"`
	Success             bool                `json:"success"`
	ResourceNameDetails ResourceNameDetails `json:"resourceNameDetails"`
}

type ResourceNameDetails struct {
	ID               int64  `json:"id"`
	ResourceName     string `json:"resourceName"`
	ResourceTypeName string `json:"resourceTypeName"`
}
