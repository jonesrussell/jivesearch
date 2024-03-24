package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Length is an instant answer
type Length struct {
	Answer
}

func (l *Length) setQuery(req *http.Request, q string) Answerer {
	l.Answer.setQuery(req, q)
	return l
}

func (l *Length) setUserAgent(req *http.Request) Answerer {
	return l
}

func (l *Length) setLanguage(lang language.Tag) Answerer {
	l.language = lang
	return l
}

func (l *Length) setType() Answerer {
	l.Type = UnitConverterType
	return l
}

func (l *Length) setRegex() Answerer {
	u := []string{
		"mile", "yard", "foot", "feet", "inch", "nautical mile",
		"ft", "in",
		// https://en.wikipedia.org/wiki/Metre (use the common units that are in boldface)
		"centimeter", "millimeter", "micrometer", "nanometer", "meter", "kilometer",
		"centimetre", "millimetre", "micrometre", "nanometre", "metre", "kilometre",
		"cm", "mm", "nm", "km", //"m", // triggers other instant answers and causes tests to fail
	}

	for i, ll := range u {
		u[i] = fmt.Sprintf(`%v[s]{0,1}\b`, ll)
	}

	lll := strings.Join(u, "|")

	t := fmt.Sprintf(`[0-9]*\s?%v to [0-9]*\s?%v`, lll, lll)

	l.regex = append(l.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	l.regex = append(l.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return l
}

func (l *Length) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	l.Solution = "length"
	return l
}

func (l *Length) tests() []test {
	d := Data{
		Type:      UnitConverterType,
		Triggered: true,
		Solution:  "length",
	}

	tests := []test{
		{
			query:    "ins to cms",
			expected: []Data{d},
		},
		{
			query:    "convert 1 meter to feet",
			expected: []Data{d},
		},
		{
			query:    "10 meters to foot",
			expected: []Data{d},
		},
		{
			query:    "cm to m",
			expected: []Data{d},
		},
		{
			query:    "inches to cm",
			expected: []Data{d},
		},
		{
			query:    "nm to feets", // why not
			expected: []Data{d},
		},
	}

	return tests
}
