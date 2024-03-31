package instant

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jonesrussell/jivesearch/instant/location"
	"github.com/jonesrussell/jivesearch/instant/weather"
	"golang.org/x/text/language"
)

// LocalWeatherType is an answer Type
const LocalWeatherType Type = "local weather"

// WeatherType is an answer Type
const WeatherType Type = "weather"

// Weather is an instant answer
type Weather struct {
	Fetcher         weather.Fetcher
	LocationFetcher location.Fetcher
	Answer
}

func (w *Weather) setQuery(r *http.Request, qv string) Answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *Weather) setUserAgent(r *http.Request) Answerer {
	return w
}

func (w *Weather) setLanguage(lang language.Tag) Answerer {
	w.language = lang
	return w
}

func (w *Weather) setType() Answerer {
	w.Type = WeatherType
	return w
}

func (w *Weather) setRegex() Answerer {
	w.regex = append(w.regex, regexp.MustCompile(`^(?P<trigger>weather|weather forecast|weather fore cast|forecast|fore cast|climate)$`))

	triggers := []string{
		"climate for", "climate",
		"forecast for", "forecast", "fore cast for", "fore cast",
		"weather forecast for", "weather forecast in", "weather forecast",
		"weather fore cast for", "weather fore cast in", "weather fore cast",
		"weather for", "weather in", "weather",
	}

	t := strings.Join(triggers, "|")
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)\s(?P<remainder>.*)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)\s(?P<trigger>%s)$`, t)))

	return w
}

func (w *Weather) solve(r *http.Request) Answerer {
	switch len(w.remainder) {
	case 0:
		return w.local(r)
	default:
		if w.remainder == "local" {
			return w.local(r)
		} else if len(w.remainder) == 5 { // U.S. zipcodes
			if z, err := strconv.Atoi(w.remainder); err == nil {
				w.Data.Solution, err = w.Fetcher.FetchByZip(z)
				if err != nil {
					w.Err = err
				}
				return w
			}
		}

		return w.city(r)
	}
}

func (w *Weather) city(_ *http.Request) *Weather {
	var err error
	w.Data.Solution, err = w.Fetcher.FetchByCity(w.remainder)
	if err != nil {
		w.Err = err
	}

	return w
}

func (w *Weather) local(r *http.Request) *Weather {
	// fetch by lat/long. On localhost this will likely give you weather for "Earth"
	w.Type = "local weather"
	ip := getIPAddress(r)

	city, err := w.LocationFetcher.Fetch(ip)
	if err != nil {
		w.Err = err
	}

	w.Data.Solution, err = w.Fetcher.FetchByLatLong(city.Location.Latitude, city.Location.Longitude, city.Location.TimeZone)
	if err != nil {
		w.Err = err
	}

	return w
}

func (w *Weather) tests() []test {
	tests := []test{
		{
			query: "local weather for",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      LocalWeatherType,
					Triggered: true,
					Solution: &weather.Weather{
						City: "Bountiful",
						Current: &weather.Instant{
							Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Low:         55,
							High:        63,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
						},
						Forecast: []*weather.Instant{
							{
								Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 97,
								Low:         84,
								High:        97,
								Wind:        3.94,
								Pressure:    888.01,
								Humidity:    14,
							},
							{
								Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 95,
								Low:         85,
								High:        95,
								Wind:        10.76,
								Pressure:    886.87,
								Humidity:    13,
							},
						},
						Provider: weather.OpenWeatherMapProvider,
						TimeZone: "America/Denver",
					},
				},
			},
		},
		{
			query: "weather for 84014",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      WeatherType,
					Triggered: true,
					Solution: &weather.Weather{
						City: "Bountiful",
						Current: &weather.Instant{
							Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Low:         55,
							High:        63,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
						},
						Forecast: []*weather.Instant{
							{
								Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 97,
								Low:         84,
								High:        97,
								Wind:        3.94,
								Pressure:    888.01,
								Humidity:    14,
							},
							{
								Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 95,
								Low:         85,
								High:        95,
								Wind:        10.76,
								Pressure:    886.87,
								Humidity:    13,
							},
						},
						Provider: weather.OpenWeatherMapProvider,
					},
				},
			},
		},
		{
			query: "weather in bogota",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      WeatherType,
					Triggered: true,
					Solution: &weather.Weather{
						City: "Bogota",
						Current: &weather.Instant{
							Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
							Code:        weather.ScatteredClouds,
							Temperature: 59,
							Low:         55,
							High:        63,
							Wind:        4.7,
							Clouds:      40,
							Rain:        0,
							Snow:        0,
							Pressure:    1014,
							Humidity:    33,
						},
						Forecast: []*weather.Instant{
							{
								Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 97,
								Low:         84,
								High:        97,
								Wind:        3.94,
								Pressure:    888.01,
								Humidity:    14,
							},
							{
								Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
								Code:        weather.Clear,
								Temperature: 95,
								Low:         85,
								High:        99,
								Wind:        10.76,
								Pressure:    886.87,
								Humidity:    13,
							},
						},
						Provider: weather.OpenWeatherMapProvider,
					},
				},
			},
		},
	}

	return tests
}
