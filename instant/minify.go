package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// MinifyType is an answer Type
const MinifyType Type = "minify"

// Minify is an instant answer
type Minify struct {
	Answer
}

func (m *Minify) setQuery(req *http.Request, q string) Answerer {
	m.Answer.setQuery(req, q)
	return m
}

func (m *Minify) setUserAgent(req *http.Request) Answerer {
	return m
}

func (m *Minify) setLanguage(lang language.Tag) Answerer {
	m.language = lang
	return m
}

func (m *Minify) setType() Answerer {
	m.Type = MinifyType
	return m
}

func (m *Minify) setRegex() Answerer {
	triggers := []string{
		"minify", "minifier", "pretty", "prettifier", "prettify",
	}

	t := strings.Join(triggers, "|")
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) .*$`, t)))
	m.regex = append(m.regex, regexp.MustCompile(fmt.Sprintf(`^.* (?P<trigger>%s)$`, t)))

	return m
}

func (m *Minify) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	return m
}

func (m *Minify) tests() []test {
	d := Data{
		Type:      MinifyType,
		Triggered: true,
	}

	tests := []test{
		{
			query:    "minify javascript",
			expected: []Data{d},
		},
		{
			query:    "pretty",
			expected: []Data{d},
		},
		{
			query:    `css prettifier`,
			expected: []Data{d},
		},
	}

	return tests
}
