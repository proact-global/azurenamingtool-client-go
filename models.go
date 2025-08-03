package azurenamingtool

// Order -
type Order struct {
	ID    int         `json:"id,omitempty"`
	Items []OrderItem `json:"items,omitempty"`
}

// OrderItem -
type OrderItem struct {
	Coffee   Coffee `json:"coffee"`
	Quantity int    `json:"quantity"`
}

// Coffee -
type Coffee struct {
	ID          int                `json:"id"`
	Name        string             `json:"name"`
	Teaser      string             `json:"teaser"`
	Collection  string             `json:"collection"`
	Origin      string             `json:"origin"`
	Color       string             `json:"color"`
	Description string             `json:"description"`
	Price       float64            `json:"price"`
	Image       string             `json:"image"`
	Ingredient  []CoffeeIngredient `json:"ingredients"`
}

// Ingredient -
type CoffeeIngredient struct {
	ID       int    `json:"ingredient_id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

// Ingredient -
type Ingredient struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Unit     string `json:"unit"`
}

// ResourceTypes -
type ResourceTypes struct {
	ID                           int    `json:"id"`
	Name                         string `json:"name"`
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
