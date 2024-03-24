package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// FrequencyType is an answer Type
const FrequencyType Type = "frequency"

var reFrequency *regexp.Regexp

// Frequency is an instant answer
type Frequency struct {
	Answer
}

func (f *Frequency) setQuery(r *http.Request, qv string) Answerer {
	f.Answer.setQuery(r, qv)
	return f
}

func (f *Frequency) setUserAgent(r *http.Request) Answerer {
	return f
}

func (f *Frequency) setLanguage(lang language.Tag) Answerer {
	f.language = lang
	return f
}

func (f *Frequency) setType() Answerer {
	f.Type = FrequencyType
	return f
}

func (f *Frequency) setRegex() Answerer {
	triggers := []string{
		"frequency of",
	}

	t := strings.Join(triggers, "|")
	f.regex = append(f.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	f.regex = append(f.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t))) // not implemented yet
	return f
}

func (f *Frequency) solve(r *http.Request) Answerer {
	var char string
	var wrd string

	matches := reFrequency.FindStringSubmatch(f.remainder)
	if len(matches) == 3 {
		char = matches[1]
		wrd = matches[2]
	}

	if char != "" && wrd != "" {
		cnt := 0
		for _, c := range wrd {
			if string(c) == char {
				cnt++
			}
		}
		f.Solution = strconv.Itoa(cnt)
	}

	return f
}

func (f *Frequency) tests() []test {
	tests := []test{
		{
			query: "a in abracadabra frequency of",
			expected: []Data{
				{
					Type:      FrequencyType,
					Triggered: true,
					Solution:  "5",
				},
			},
		},
		{
			query: "frequency of a in abracadabra",
			expected: []Data{
				{
					Type:      FrequencyType,
					Triggered: true,
					Solution:  "5",
				},
			},
		},
		{
			query: "frequency of o in cooler",
			expected: []Data{
				{
					Type:      FrequencyType,
					Triggered: true,
					Solution:  "2",
				},
			},
		},
		{
			query: "frequency of s in jimi hendrix",
			expected: []Data{
				{
					Type:      FrequencyType,
					Triggered: true,
					Solution:  "0",
				},
			},
		},
		{
			query: "frequency of e in fred astaire",
			expected: []Data{
				{
					Type:      FrequencyType,
					Triggered: true,
					Solution:  "2",
				},
			},
		},
	}

	return tests

}

func init() {
	reFrequency = regexp.MustCompile(`^(.*?) in (.+)`)
}
