// Package parcel shows package status for UPS, Fedex and others
package parcel

import "time"

// Fetcher retrieves package info from the UPS API
type Fetcher interface {
	Fetch(number string) (Response, error)
}

// Response is a standardized response for tracking packages
type Response struct {
	TrackingNumber string   `json:"tracking_number"`
	Updates        []Update `json:"updates"`
	Expected
	URL string `json:"url"`
}

// Expected is the expected delivery date and time
type Expected struct {
	Delivery string    `json:"delivery"`
	Date     time.Time `json:"expected"`
}

// Update is a single delivery event for a package
type Update struct {
	DateTime time.Time `json:"date_time"`
	Location `json:"location"`
	Status   string `json:"status"`
}

// Location is a location of an Update
type Location struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
}
