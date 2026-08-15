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
	ResourceEnvironment string `json:"resourceEnvironment"`
	ResourceFunction    string `json:"resourceFunction"`
	ResourceInstance    string `json:"resourceInstance"`
	ResourceLocation    string `json:"resourceLocation"`
	ResourceOrg         string `json:"resourceOrg"`
	ResourceType        string `json:"resourceType"`

	// ResourceProjAppSvc and ResourceUnitDept are components the naming tool
	// supports and includes in a name when enabled. They were previously absent
	// here, so a deployment using either could not be driven through this client.
	ResourceProjAppSvc string `json:"resourceProjAppSvc,omitempty"`
	ResourceUnitDept   string `json:"resourceUnitDept,omitempty"`

	CustomComponents GenerateNameRequestCustomComponents `json:"customComponents"`
}

// GenerateNameRequestCustomComponents holds the custom naming components for a
// GenerateNameRequest. The map key is the component name (e.g. "application",
// "subnet_tier", "subnet_instance") and the value is its registered short name.
type GenerateNameRequestCustomComponents = map[string]string

type GenerateNameResponse struct {
	ResourceName        string              `json:"resourceName"`
	Message             string              `json:"message"`
	Success             bool                `json:"success"`
	ResourceNameDetails ResourceNameDetails `json:"resourceNameDetails"`
	ValidationMetadata  ValidationMetadata  `json:"validationMetadata"`
}

type ResourceNameDetails struct {
	ID               int64      `json:"id"`
	ResourceName     string     `json:"resourceName"`
	ResourceTypeName string     `json:"resourceTypeName"`
	Components       [][]string `json:"components"`
}

type DeleteGeneratedNameRequest struct {
	ID int64 `json:"id"`
}

// V2 API wrapper types

type ApiResponse[T any] struct {
	Success  bool        `json:"success"`
	Data     T           `json:"data"`
	Error    *ApiError   `json:"error"`
	Metadata ApiMetadata `json:"metadata"`
}

type ApiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Target  string `json:"target"`
}

type ApiMetadata struct {
	CorrelationID string `json:"correlationId"`
	Timestamp     string `json:"timestamp"`
	Message       string `json:"message"`
}

type ValidationMetadata struct {
	ValidationPerformed bool   `json:"validationPerformed"`
	ExistsInAzure       bool   `json:"existsInAzure"`
	OriginalName        string `json:"originalName"`
	IncrementAttempts   int    `json:"incrementAttempts"`
	ValidationWarning   string `json:"validationWarning"`
}
