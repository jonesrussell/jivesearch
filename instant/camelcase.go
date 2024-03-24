package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// CamelCaseType is an answer Type
const CamelCaseType Type = "camelcase"

// CamelCase is an instant answer
type CamelCase struct {
	Answer
}

func (c *CamelCase) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *CamelCase) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *CamelCase) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *CamelCase) setType() Answerer {
	c.Type = CamelCaseType
	return c
}

func (c *CamelCase) setRegex() Answerer {
	triggers := []string{
		"camelcase",
		"camel case",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return c
}

func (c *CamelCase) solve(r *http.Request) Answerer {
	titled := []string{}
	for _, w := range strings.Fields(c.remainder) {
		titled = append(titled, strings.Title(w))
	}

	c.Solution = strings.Join(titled, "")

	return c
}

func (c *CamelCase) tests() []test {
	tests := []test{
		{
			query: "camelcase metallica rocks",
			expected: []Data{
				{
					Type:      CamelCaseType,
					Triggered: true,
					Solution:  "MetallicaRocks",
				},
			},
		},
		{
			query: "aliCE in chAins Is better camel case",
			expected: []Data{
				{
					Type:      CamelCaseType,
					Triggered: true,
					Solution:  "AliceInChainsIsBetter",
				},
			},
		},
		{
			query: "camel case O'doyle ruLES",
			expected: []Data{
				{
					Type:      CamelCaseType,
					Triggered: true,
					Solution:  "O'DoyleRules",
				},
			},
		},
	}

	return tests
}
