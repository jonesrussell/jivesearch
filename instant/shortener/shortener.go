// Package shortener shortens urls
package shortener

import "net/url"

// Service is a url shortening service
type Service interface {
	Shorten(u *url.URL) (*Response, error)
}

type provider string

// Response is a link shortener response
type Response struct {
	Original *url.URL
	Short    *url.URL
	Provider provider
}
