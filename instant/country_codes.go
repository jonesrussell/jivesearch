package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/pariz/gountries"
	"golang.org/x/text/language"
)

// CountryCodeType is an answer Type
const CountryCodeType Type = "country code"

// CountryCode is an instant answer
type CountryCode struct {
	Answer
}

// CountryCodeResponse is a response to the instant answer
type CountryCodeResponse struct {
	Format   string
	Country  string
	Solution string
}

// ISO3166 is the ISO 3166-1 alpha-2 country code format
const ISO3166 = "ISO 3166-1 alpha-2"

func (c *CountryCode) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *CountryCode) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *CountryCode) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *CountryCode) setType() Answerer {
	c.Type = CountryCodeType
	return c
}

func (c *CountryCode) setRegex() Answerer {
	triggers := []string{
		"country code", "iso code", "iso 3166", "iso",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return c
}

func (c *CountryCode) solve(r *http.Request) Answerer {
	query := gountries.New()

	country, err := query.FindCountryByName(c.remainder)
	if err != nil {
		country, err = query.FindCountryByAlpha(c.remainder)
		if err != nil {
			c.Err = err
			return c
		}
	}

	c.Solution = CountryCodeResponse{
		Format:   ISO3166,
		Country:  country.Name.Common,
		Solution: country.Alpha2,
	}

	return c
}

func (c *CountryCode) tests() []test {
	tests := []test{
		{
			query: "country code united states",
			expected: []Data{
				{
					Type:      CountryCodeType,
					Triggered: true,
					Solution: CountryCodeResponse{
						Format:   ISO3166,
						Country:  "United States",
						Solution: "US",
					},
				},
			},
		},
		{
			query: "iso DE",
			expected: []Data{
				{
					Type:      CountryCodeType,
					Triggered: true,
					Solution: CountryCodeResponse{
						Format:   ISO3166,
						Country:  "Germany",
						Solution: "DE",
					},
				},
			},
		},
		{
			query: "iso code denmark",
			expected: []Data{
				{
					Type:      CountryCodeType,
					Triggered: true,
					Solution: CountryCodeResponse{
						Format:   ISO3166,
						Country:  "Denmark",
						Solution: "DK",
					},
				},
			},
		},
		{
			query: "iso 3166 sweden",
			expected: []Data{
				{
					Type:      CountryCodeType,
					Triggered: true,
					Solution: CountryCodeResponse{
						Format:   ISO3166,
						Country:  "Sweden",
						Solution: "SE",
					},
				},
			},
		},
		{
			query: "italy country code",
			expected: []Data{
				{
					Type:      CountryCodeType,
					Triggered: true,
					Solution: CountryCodeResponse{
						Format:   ISO3166,
						Country:  "Italy",
						Solution: "IT",
					},
				},
			},
		},
	}

	return tests
}
