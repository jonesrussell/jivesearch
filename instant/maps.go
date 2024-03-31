package instant

import (
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/jonesrussell/jivesearch/instant/location"
	"golang.org/x/text/language"
)

// MapsType is an answer Type
const MapsType Type = "maps"

// Maps is an instant answer
type Maps struct {
	LocationFetcher location.Fetcher
	Answer
}

// Map is a instant answer response
type Map struct {
	location.Location
	Directions  bool
	Origin      string
	Destination string
}

func (m *Maps) setQuery(req *http.Request, q string) Answerer {
	m.Answer.setQuery(req, q)
	return m
}

func (m *Maps) setUserAgent(req *http.Request) Answerer {
	return m
}

func (m *Maps) setLanguage(lang language.Tag) Answerer {
	m.language = lang
	return m
}

func (m *Maps) setType() Answerer {
	m.Type = MapsType
	return m
}

func (m *Maps) setRegex() Answerer {
	m.regex = append(m.regex, regexp.MustCompile(`^(directions|map) to (?P<end>.*)$`))
	m.regex = append(m.regex, regexp.MustCompile(`^(directions|map) (?P<start>.*) to (?P<end>.*)$`))
	m.regex = append(m.regex, regexp.MustCompile(`^(?P<start>.*) to (?P<end>.*) (directions|map)$`))
	m.regex = append(m.regex, regexp.MustCompile(`^(?P<end>.*) (directions|map)$`))
	m.regex = append(m.regex, regexp.MustCompile(`^(directions|map) (?P<end>.*)$`))

	triggers := []string{
		"map", "maps", "direction", "directions",
	}

	t := strings.Join(triggers, "|")
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) .*$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^.* (?P<trigger>%s)$`, t)))

	return m
}

func (m *Maps) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	// Note: "Boston map" needs directions as we are unable to simply get the Lat/Long
	// "Boston", on the other hand, gets the Lat/Long from Wikidata and is triggered there.
	ip := getIPAddress(r)

	city, err := m.LocationFetcher.Fetch(ip)
	if err != nil {
		m.Err = err
	}

	mm := Map{
		Location: location.Location{
			Latitude:  city.Location.Latitude,
			Longitude: city.Location.Longitude,
		},
	}

	if m.triggerWord == "direction" || m.triggerWord == "directions" {
		mm.Directions = true
	}

	mm.Origin = m.remainderM["start"]
	var ok bool

	if mm.Destination, ok = m.remainderM["end"]; ok {
		mm.Directions = true
		if mm.Origin == "" { // if no starting point then use their current location
			if c, ok := city.City.Names["en"]; ok {
				if len(city.Subdivisions) > 0 {
					if s, ok := city.Subdivisions[0].Names["en"]; ok {
						mm.Origin = fmt.Sprintf("%v, %v", c, s)
					}
				}
			}
		}
	}

	m.Data.Solution = mm

	return m
}

func (m *Maps) tests() []test {
	tests := []test{
		{
			query: "map",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location: location.Location{Latitude: 12, Longitude: 18},
					},
				},
			},
		},
		{
			query: "map new york city",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location:    location.Location{Latitude: 12, Longitude: 18},
						Directions:  true,
						Origin:      "Someville, SomeState",
						Destination: "new york city",
					},
				},
			},
		},
		{
			query: "directions",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location:   location.Location{Latitude: 12, Longitude: 18},
						Directions: true,
					},
				},
			},
		},
		{
			query: "directions to new york city",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location:    location.Location{Latitude: 12, Longitude: 18},
						Directions:  true,
						Origin:      "Someville, SomeState",
						Destination: "new york city",
					},
				},
			},
		},
		{
			query: "new york city directions",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location:    location.Location{Latitude: 12, Longitude: 18},
						Directions:  true,
						Origin:      "Someville, SomeState",
						Destination: "new york city",
					},
				},
			},
		},
		{
			query: "san francisco to ohio directions",
			ip:    net.ParseIP("161.59.224.138"),
			expected: []Data{
				{
					Type:      MapsType,
					Triggered: true,
					Solution: Map{
						Location:    location.Location{Latitude: 12, Longitude: 18},
						Directions:  true,
						Origin:      "san francisco",
						Destination: "ohio",
					},
				},
			},
		},
	}

	return tests
}
