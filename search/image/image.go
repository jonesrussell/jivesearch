package image

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Image is a link to an image
type Image struct {
	ID     string  `json:"id"`
	Domain string  `json:"domain"`
	Alt    string  `json:"alt,omitempty"`
	NSFW   float64 `json:"nsfw_score,omitempty"`
	Width  int     `json:"width,omitempty"`
	Height int     `json:"height,omitempty"`
	EXIF
	Classification map[string]float64 `json:"classification,omitempty"`
	MIME           string             `json:"mime,omitempty"`
	Crawled        string             `json:"crawled,omitempty"`
	Base64         string             `json:"base64,omitempty"`
}

// EXIF is the metadata of an image
// What other fields do we need???
type EXIF struct {
	Copyright string `json:"copyright,omitempty"`
}

// Fetcher outlines the methods used to retrieve the image results
type Fetcher interface {
	Fetch(q string, safe bool, number int, offset int) (*Results, error)
}

// Provider is an image source
type Provider string

// Results are the image results from a query
type Results struct {
	Provider   Provider
	Count      int64    `json:"count"`
	Page       string   `json:"page"`
	Previous   string   `json:"previous"`
	Next       string   `json:"next"`
	Last       string   `json:"last"`
	Pagination []string `json:"pagination"`
	Images     []*Image `json:"images"`
}

var errInvalidURL = fmt.Errorf("invalid url")

// New creates a new *Image and validates the url
func New(src string) (*Image, error) {
	u, err := url.Parse(src)
	if err != nil {
		return nil, err
	}

	if u.String() == "" {
		return nil, errInvalidURL
	}

	dom, err := publicsuffix.EffectiveTLDPlusOne(u.Host)
	if err != nil {
		return nil, errInvalidURL
	}

	return &Image{ID: u.String(), Domain: dom}, nil
}

// SimplifyMIME strips unnecessary info from MIME
// image/jpg -> jpg
func (i *Image) SimplifyMIME() *Image {
	s := strings.Split(i.MIME, "/")
	if len(s) > 1 {
		i.MIME = s[1]
	}

	return i
}
