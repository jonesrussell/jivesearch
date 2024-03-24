package nutrition

import "encoding/json"

// Fetcher outlines methods to get nutrition info for a food
type Fetcher interface {
	Lookup(query string) ([]*ItemResponse, error)
	Fetch([]string) (*Response, error)
}

type provider string

// ItemResponse is a lookup item response
type ItemResponse struct {
	Name         string `json:"name"`
	NDBNO        string `json:"ndbno"`
	Manufacturer string `json:"manu"`
}

// Response is a nutrition response
type Response struct {
	Foods    []Food
	Provider provider
}

// Food is a single food item
type Food struct {
	Name        string
	FoodGroup   string
	Corporation string
	Nutrients   []Nutrient
}

// Nutrient is a single nutrient
type Nutrient struct {
	ID       json.Number
	Name     string
	Unit     string
	Value    json.Number
	Measures []Measure
}

// Measure is an alternative measurement
type Measure struct {
	Label      string
	Equivalent float64
	Units      string
	Quantity   float64
	Value      json.Number
}
