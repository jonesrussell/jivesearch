package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/jivesearch/jivesearch/instant/econ"
	"github.com/jivesearch/jivesearch/instant/econ/gdp"
	"github.com/pariz/gountries"
	"golang.org/x/text/language"
)

// GDP is an instant answer
type GDP struct {
	GDPFetcher gdp.Fetcher
	Answer
}

// GDPResponse is an instant answer response
type GDPResponse struct {
	Country string
	*gdp.Response
}

// ErrInvalidCountry indicates a country is not valid
var ErrInvalidCountry error

func (g *GDP) setQuery(r *http.Request, qv string) Answerer {
	g.Answer.setQuery(r, qv)
	return g
}

func (g *GDP) setUserAgent(r *http.Request) Answerer {
	return g
}

func (g *GDP) setLanguage(lang language.Tag) Answerer {
	g.language = lang
	return g
}

func (g *GDP) setType() Answerer {
	g.Type = "gdp"
	return g
}

func (g *GDP) setRegex() Answerer {
	for _, w := range []string{"gdp", "gross domestic product"} {
		regexes := []string{
			fmt.Sprintf(`^(?P<country>.*) %v$`, w),
			fmt.Sprintf(`^(?P<country>.*) %v of$`, w),
			fmt.Sprintf(`^%v of (?P<country>.*)$`, w),
			fmt.Sprintf(`^%v (?P<country>.*)$`, w),
		}

		for _, rgx := range regexes {
			g.regex = append(g.regex, regexp.MustCompile(rgx))
		}
	}

	return g
}

func (g *GDP) solve(r *http.Request) Answerer {
	c, ok := g.remainderM["country"]
	if !ok {
		g.Err = ErrInvalidCountry
		return g
	}

	// is it a valid country?
	query := gountries.New()

	country, err := query.FindCountryByName(c)
	if err != nil {
		country, err = query.FindCountryByAlpha(c)
		if err != nil {
			g.Err = err
			return g
		}
	}

	alpha := country.Alpha2

	resp := &GDPResponse{
		Country: country.Name.Common,
	}

	n := time.Now().Year()
	start := n - 50 // 50 years seems to be the max allowed

	resp.Response, err = g.GDPFetcher.Fetch(alpha, time.Date(start, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(n, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		g.Err = err
		return g
	}

	resp.Response.Sort()

	g.Data.Solution = resp
	return g
}

func (g *GDP) setCache() Answerer {
	g.Cache = true
	return g
}

func (g *GDP) tests() []test {
	typ := "gdp"

	tests := []test{
		{
			query: "Italy gdp",
			expected: []Data{
				{
					Type:      typ,
					Triggered: true,
					Solution: &GDPResponse{
						Country: "Italy",
						Response: &gdp.Response{
							History: []gdp.Instant{
								{
									Date:  time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 4,
								},
								{
									Date:  time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 2,
								},
								{
									Date:  time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC),
									Value: 18,
								},
							},
							Provider: econ.TheWorldBankProvider,
						},
					},
					Cache: true,
				},
			},
		},
	}

	return tests
}
