package location

import (
	"net"

	geoip2 "github.com/oschwald/geoip2-golang"
)

// MaxMind is a location data provider
// For install instructions: https://dev.maxmind.com/geoip/geoipupdate/
type MaxMind struct {
	DB string // location of MaxMind database
}

var open = func(loc string) (maxMinder, error) {
	return geoip2.Open(loc)
}

// Fetch gets geolocation data from an IP Address
func (m *MaxMind) Fetch(ip net.IP) (*City, error) {
	db, err := open(m.DB)
	if err != nil {
		return nil, err
	}

	defer db.Close()

	c, err := db.City(ip)
	if err != nil {
		return nil, err
	}

	// since xml marshaller cannot marshal map[string]string we have to create
	// our own City struct and fill it in manually...converting to from
	// map[string]string to xmlMap is not possible
	// https://stackoverflow.com/questions/54301474/how-do-i-implement-custom-struct-tags-and-be-able-to-xml-encode-a-map
	var city = &City{}
	city.City.GeoNameID = c.City.GeoNameID
	city.City.Names = c.City.Names
	city.Continent.Code = c.Continent.Code
	city.Continent.GeoNameID = c.Continent.GeoNameID
	city.Continent.Names = c.Continent.Names
	city.Country.GeoNameID = c.Country.GeoNameID
	city.Country.IsInEuropeanUnion = c.Country.IsInEuropeanUnion
	city.Country.IsoCode = c.Country.IsoCode
	city.Country.Names = c.Country.Names
	city.Location.AccuracyRadius = c.Location.AccuracyRadius
	city.Location.Latitude = c.Location.Latitude
	city.Location.Longitude = c.Location.Longitude
	city.Location.MetroCode = c.Location.MetroCode
	city.Location.TimeZone = c.Location.TimeZone
	city.Postal.Code = c.Postal.Code
	city.RegisteredCountry.GeoNameID = c.RegisteredCountry.GeoNameID
	city.RegisteredCountry.IsInEuropeanUnion = c.RegisteredCountry.IsInEuropeanUnion
	city.RegisteredCountry.IsoCode = c.RegisteredCountry.IsoCode
	city.RegisteredCountry.Names = c.RegisteredCountry.Names
	city.RepresentedCountry.GeoNameID = c.RepresentedCountry.GeoNameID
	city.RepresentedCountry.IsInEuropeanUnion = c.RepresentedCountry.IsInEuropeanUnion
	city.RepresentedCountry.IsoCode = c.RepresentedCountry.IsoCode
	city.RepresentedCountry.Names = c.RepresentedCountry.Names
	city.RepresentedCountry.Type = c.RepresentedCountry.Type
	for _, s := range c.Subdivisions {
		sub := Subdivision{
			GeoNameID: s.GeoNameID,
			IsoCode:   s.IsoCode,
			Names:     s.Names,
		}
		city.Subdivisions = append(city.Subdivisions, sub)
	}
	city.Traits.IsAnonymousProxy = c.Traits.IsAnonymousProxy
	city.Traits.IsSatelliteProvider = c.Traits.IsSatelliteProvider

	return city, err
}

type maxMinder interface {
	Close() error
	City(ipAddress net.IP) (*geoip2.City, error)
}
