// Package breach checks for data breaches for a given account
package breach

import (
	"sort"
	"time"
)

// Fetcher implements methods to check for security breaches for an account
type Fetcher interface {
	Fetch(account string) (*Response, error)
}

type provider string

// Response is a currency response
type Response struct {
	Account  string
	Breaches []Breach
	Provider provider
}

// Breach is a single breach
type Breach struct {
	Name        string
	Domain      string
	Date        time.Time
	Count       int
	Description string
	Items       []string
}

// New creates a new breach Response
func New(account string, provider provider) *Response {
	return &Response{
		Account:  account,
		Provider: provider,
	}
}

// Sort will organize the History in ascending order by date
func (r *Response) Sort() {
	sort.Slice(r.Breaches, func(i, j int) bool {
		return r.Breaches[i].Date.Before(r.Breaches[j].Date)
	})
}
