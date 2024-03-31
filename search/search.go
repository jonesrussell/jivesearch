// Package search provides the core search results.
package search

import (
	"math"
	"strconv"

	"github.com/jonesrussell/jivesearch/search/document"
	"golang.org/x/text/language"
)

// Fetcher outlines the methods used to retrieve the core search results
type Fetcher interface {
	Fetch(q string, s Filter, lang language.Tag, region language.Region, number int, offset int) (*Results, error)
}

// Provider is a search provider
type Provider string

// Filter is the safe search settings
type Filter string

// Strict indicates the strongest safe search settings
var Strict Filter = "strict"

// Off indicates the weakest safe search settings
var Off Filter = "off"

// Moderate indicates a moderate safe search setting
var Moderate Filter = "moderate"

// Results are the core search results from a query
type Results struct {
	Provider   Provider             `json:"-"`
	Count      int64                `json:"-"`
	Page       string               `json:"-"`
	Previous   string               `json:"-"`
	Next       string               `json:"next"`
	Last       string               `json:"-"`
	Pagination []string             `json:"-"`
	Documents  []*document.Document `json:"documents"`
	Err        error
}

// AddPagination adds pagination to the search results
func (r *Results) AddPagination(number, page int) *Results {
	r.Pagination = []string{}
	r.Page = strconv.Itoa(page)
	if page > 1 {
		r.Previous = strconv.Itoa(page - 1)
	}

	min, max := 1, int(math.Ceil(float64(r.Count)/float64(number))) // round up
	if page > max {
		page = max
	}

	if max > page {
		r.Next = strconv.Itoa(page + 1)
	}

	if page > 6 {
		min = page - 5
		tmp := max
		for i := 0; i < 5; i++ {
			if tmp < max {
				max = page + i
			}
		}
	}

	for i := min; i <= max; i++ {
		if len(r.Pagination) < 10 {
			r.Pagination = append(r.Pagination, strconv.Itoa(i))
		}
	}

	return r
}
