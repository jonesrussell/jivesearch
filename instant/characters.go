package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// CharactersType is an answer Type
const CharactersType Type = "characters"

// Characters is an instant answer
type Characters struct {
	Answer
}

func (c *Characters) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Characters) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *Characters) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *Characters) setType() Answerer {
	c.Type = CharactersType
	return c
}

func (c *Characters) setRegex() Answerer {
	triggers := []string{
		"number of characters in", "number of characters",
		"number of chars in", "number of chars",
		"char count of", "char count",
		"chars count of", "chars count",
		"character count of", "character count",
		"characters count of", "characters count",
		"length in chars", "length in characters",
	}

	t := strings.Join(triggers, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return c
}

func (c *Characters) solve(r *http.Request) Answerer {
	for _, ch := range []string{`"`, `'`} {
		c.remainder = strings.TrimPrefix(c.remainder, ch)
		c.remainder = strings.TrimSuffix(c.remainder, ch)
	}

	c.Solution = strconv.Itoa(len(c.remainder))

	return c
}

func (c *Characters) tests() []test {
	tests := []test{
		{
			query: `number of chars in "Jimi Hendrix"`,
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "12",
				},
			},
		},
		{
			query: "number of chars   in Pink   Floyd",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "10",
				},
			},
		},
		{
			query: "Bob Dylan   number of characters in",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "9",
				},
			},
		},
		{
			query: "number of characters Janis   Joplin",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "12",
				},
			},
		},
		{
			query: "char count Led Zeppelin",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "12",
				},
			},
		},
		{
			query: "char count of ' 87 '",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "4",
				},
			},
		},
		{
			query: "they're chars count",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "7",
				},
			},
		},
		{
			query: "chars count of something",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "9",
				},
			},
		},
		{
			query: "Another something chars count of",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "17",
				},
			},
		},
		{
			query: "1234567 character count",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "7",
				},
			},
		},
		{
			query: "character count of house of cards",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "14",
				},
			},
		},
		{
			query: "characters count 50 cent",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "7",
				},
			},
		},
		{
			query: "characters count of 1 dollar",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "8",
				},
			},
		},
		{
			query: "char count equity",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "6",
				},
			},
		},
		{
			query: "characters count seal",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "4",
				},
			},
		},
		{
			query: "length in chars lion",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "4",
				},
			},
		},
		{
			query: "length in characters mountain",
			expected: []Data{
				{
					Type:      CharactersType,
					Triggered: true,
					Solution:  "8",
				},
			},
		},
	}

	return tests
}
