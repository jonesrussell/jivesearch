// Package timezone retrieves timezone info from lat/lon coordinates
package timezone

// Fetcher lays out methods to retrieve timezone data
type Fetcher interface {
	Fetch(lat, lon float64) (string, error)
}
