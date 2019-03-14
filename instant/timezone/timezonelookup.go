package timezone

import (
	timezone "github.com/evanoberholster/timezoneLookup"
)

// TZLookup holds settings for timezoneLookup
type TZLookup struct {
	TZ timezone.TimezoneInterface
}

// Fetch retrieves the timezone from a local database
func (t *TZLookup) Fetch(lat, lon float64) (string, error) {
	return t.TZ.Query(
		timezone.Coord{Lat: float32(lon), Lon: float32(lat)}, // is this backwards or just me???
	)
}
