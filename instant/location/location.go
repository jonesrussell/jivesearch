// Package location fetches geolocation data
package location

import (
	"encoding/xml"
	"net"
)

// Location is the lat/long of a place
type Location struct {
	Latitude  float64
	Longitude float64
}

// Fetcher retrieves stock quotes
type Fetcher interface {
	Fetch(ip net.IP) (*City, error)
}

type xmlMap map[string]string
type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// City is like geoip2.City but adds json tags
type City struct {
	City struct {
		GeoNameID uint   `maxminddb:"geoname_id" json:"geoname_id" xml:"geoname_id"`
		Names     xmlMap `maxminddb:"names" json:"names" xml:"names"`
	} `maxminddb:"city" json:"city" xml:"city"`
	Continent struct {
		Code      string `maxminddb:"code" json:"code" xml:"code"`
		GeoNameID uint   `maxminddb:"geoname_id" json:"geoname_id" xml:"geoname_id"`
		Names     xmlMap `maxminddb:"names" json:"names" xml:"names"`
	} `maxminddb:"continent" json:"continent" xml:"continent"`
	Country struct {
		GeoNameID         uint   `maxminddb:"geoname_id" json:"geoname_id" xml:"geoname_id"`
		IsInEuropeanUnion bool   `maxminddb:"is_in_european_union" json:"is_eu" xml:"is_eu"`
		IsoCode           string `maxminddb:"iso_code" json:"iso_code" xml:"iso_code"`
		Names             xmlMap `maxminddb:"names" json:"names" xml:"names"`
	} `maxminddb:"country" json:"country" xml:"country"`
	Location struct {
		AccuracyRadius uint16  `maxminddb:"accuracy_radius" json:"accuracy_radius" xml:"accuracy_radius"`
		Latitude       float64 `maxminddb:"latitude" json:"latitude" xml:"latitude"`
		Longitude      float64 `maxminddb:"longitude" json:"longitude" xml:"longitude"`
		MetroCode      uint    `maxminddb:"metro_code" json:"metro_code" xml:"metro_code"`
		TimeZone       string  `maxminddb:"time_zone" json:"time_zone" xml:"time_zone"`
	} `maxminddb:"location" json:"location" xml:"location"`
	Postal struct {
		Code string `maxminddb:"code" json:"code" xml:"code"`
	} `maxminddb:"postal" json:"postal" xml:"postal"`
	RegisteredCountry struct {
		GeoNameID         uint   `maxminddb:"geoname_id" json:"geoname_id"  xml:"geoname_id"`
		IsInEuropeanUnion bool   `maxminddb:"is_in_european_union" json:"is_eu" xml:"is_eu"`
		IsoCode           string `maxminddb:"iso_code" json:"iso_code" xml:"iso_code"`
		Names             xmlMap `maxminddb:"names" json:"names" xml:"names"`
	} `maxminddb:"registered_country" json:"registered_country" xml:"registered_country"`
	RepresentedCountry struct {
		GeoNameID         uint   `maxminddb:"geoname_id" json:"geoname_id"  xml:"geoname_id"`
		IsInEuropeanUnion bool   `maxminddb:"is_in_european_union" json:"is_eu" xml:"is_eu"`
		IsoCode           string `maxminddb:"iso_code" json:"iso_code" xml:"iso_code"`
		Names             xmlMap `maxminddb:"names" json:"names" xml:"names"`
		Type              string `maxminddb:"type"  json:"type" xml:"type"`
	} `maxminddb:"represented_country"  json:"represented_country" xml:"represented_country"`
	Subdivisions []Subdivision `maxminddb:"subdivisions" json:"subdivisions" xml:"subdivisions"`
	Traits       struct {
		IsAnonymousProxy    bool `maxminddb:"is_anonymous_proxy" json:"is_anonymous_proxy" xml:"is_anonymous_proxy"`
		IsSatelliteProvider bool `maxminddb:"is_satellite_provider" json:"is_satellite_provider" xml:"is_satellite_provider"`
	} `maxminddb:"traits" json:"traits" xml:"traits"`
}

// Subdivision adds json tags to maxmind
type Subdivision struct {
	GeoNameID uint   `maxminddb:"geoname_id" json:"geoname_id" xml:"geoname_id"`
	IsoCode   string `maxminddb:"iso_code" json:"iso_code" xml:"iso_code"`
	Names     xmlMap `maxminddb:"names" json:"names" xml:"names"`
}

// Custom XML marshalling for maps since default will give error when trying to XML marshal a map
func (m xmlMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m {
		if err := e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v}); err != nil {
			return err
		}
	}

	return e.EncodeToken(start.End())
}
