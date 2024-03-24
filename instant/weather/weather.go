// Package weather fetches weather data
package weather

import "time"

// Fetcher retrieves the current and forecasted weather
// How to get the timezone of a zipcode?
// http://download.geonames.org/export/zip/? gives lat/long of a zipcode then
// https://stackoverflow.com/a/16086964/522962 for the timezone.
type Fetcher interface {
	FetchByCity(city string) (*Weather, error)
	FetchByLatLong(lat, long float64, timeZone string) (*Weather, error)
	FetchByZip(zip int) (*Weather, error)
}

type provider string

// Weather includes the urrent and forecasted weather
type Weather struct {
	City     string
	TimeZone string
	Current  *Instant
	Forecast []*Instant
	Provider provider
	//Updated   time.Time // requires getting the timezone of the user...https://github.com/bradfitz/latlong ???
}

// Instant is the weather for a specified time
type Instant struct {
	Date        time.Time
	Code        Description
	Temperature int
	Low         int
	High        int
	Wind        float64
	Clouds      float64
	Rain        float64
	Snow        float64
	Pressure    float64
	Humidity    float64
}

// Description is a standardized weather description
type Description string

// Clear indicates clear skies
const Clear Description = "Clear"

// LightClouds indicate some clouds
const LightClouds Description = "Light Clouds"

// ScatteredClouds indicates scattered clouds
const ScatteredClouds Description = "Scattered Clouds"

// OvercastClouds indicates heavy clouds
const OvercastClouds Description = "Overcast"

// Extreme indicates an extreme event
const Extreme Description = "Extreme"

// Rain indicates rain
const Rain Description = "Rain"

// Snow indicates snow
const Snow Description = "Snow"

// ThunderStorm indicates a thunderstorm
const ThunderStorm Description = "Thunderstorm"

// Windy indicates winds
const Windy Description = "Windy"
