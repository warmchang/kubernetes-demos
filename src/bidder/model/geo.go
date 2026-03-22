package model

// GeoTarget represents geographic targeting configuration
type GeoTarget struct {
	Country string `json:"country"`
	Region  string `json:"region,omitempty"`
	City    string `json:"city,omitempty"`
	ZipCode string `json:"zip_code,omitempty"`
}

// GeoModifier holds bid adjustment percentages per geo
type GeoModifier struct {
	Target     GeoTarget `json:"target"`
	Multiplier float64   `json:"multiplier"` // 1.0 = no change, 1.5 = +50%
}
