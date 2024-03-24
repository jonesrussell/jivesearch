// Package discography fetches artist discography
package discography

import (
	"net/url"
	"time"
)

// Fetcher fetches cover art
type Fetcher interface {
	Fetch(artist string) ([]Album, error)
}

// Album is an individual album
type Album struct {
	Name      string
	Published time.Time
	Image
}

// Image is an album's image
type Image struct {
	ID  string
	URL *url.URL
}
