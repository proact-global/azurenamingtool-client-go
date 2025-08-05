package azurenamingtool

// ResourceTypes -
type ResourceTypes struct {
	ID                           int    `json:"id"`
	Resource                     string `json:"resource"`
	Optional                     string `json:"optional"`
	Exclude                      string `json:"exclude"`
	Property                     string `json:"property"`
	ShortName                    string `json:"short_name"`
	Scope                        string `json:"scope"`
	LengthMin                    string `json:"length_min"`
	LengthMax                    string `json:"length_max"`
	ValidText                    string `json:"valid_text"`
	InvalidText                  string `json:"invalid_text"`
	InvalidCharacters            string `json:"invalid_characters"`
	InvalidCharactersStart       string `json:"invalid_characters_start"`
	InvalidCharactersEnd         string `json:"invalid_characters_end"`
	InvalidCharactersConsecutive string `json:"invalid_characters_consecutive"`
	Regx                         string `json:"regx"`
	StaticValues                 string `json:"static_values"`
	Enabled                      bool   `json:"enabled"`
	ApplyDelimiter               bool   `json:"apply_delimiter"`
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
