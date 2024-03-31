// Package gdp retrieves gdp data
package gdp

import (
	"sort"
	"time"

	"github.com/jonesrussell/jivesearch/instant/econ"
)

// Provider indicates the source of the data
type Provider string

// Instant is the gdp for a point in time
type Instant struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// Response is a gdp response
type Response struct {
	History []Instant
	econ.Provider
}

// Fetcher outlines methods to retrieve gdp data
type Fetcher interface {
	Fetch(country string, start time.Time, end time.Time) (*Response, error)
}

// Sort will organize the History in ascending order by date
func (r *Response) Sort() {
	sort.Slice(r.History, func(i, j int) bool {
		return r.History[i].Date.Before(r.History[j].Date)
	})
}
