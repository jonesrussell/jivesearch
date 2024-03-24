// Package population retrieves population data
package population

import (
	"sort"
	"time"

	"github.com/jivesearch/jivesearch/instant/econ"
)

// Instant is the population for a point in time
type Instant struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// Response is a population response
type Response struct {
	History []Instant
	econ.Provider
}

// Fetcher outlines methods to retrieve population data
type Fetcher interface {
	Fetch(country string, start time.Time, end time.Time) (*Response, error)
}

// Sort will organize the History in ascending order by date
func (r *Response) Sort() {
	sort.Slice(r.History, func(i, j int) bool {
		return r.History[i].Date.Before(r.History[j].Date)
	})
}
