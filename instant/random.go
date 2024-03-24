package instant

import (
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/text/language"
)

// RandomType is an answer Type
const RandomType Type = "random"

// Random is an instant answer
type Random struct {
	Answer
}

var reRandom *regexp.Regexp

func (r *Random) setQuery(req *http.Request, qv string) Answerer {
	r.Answer.setQuery(req, qv)
	return r
}

func (r *Random) setUserAgent(req *http.Request) Answerer {
	return r
}

func (r *Random) setLanguage(lang language.Tag) Answerer {
	r.language = lang
	return r
}

func (r *Random) setType() Answerer {
	r.Type = RandomType
	return r
}

func (r *Random) setRegex() Answerer {
	triggers := []string{
		"random number", "random number between",
	}

	t := strings.Join(triggers, "|")
	r.regex = append(r.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	r.regex = append(r.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return r
}

func (r *Random) solve(req *http.Request) Answerer {
	matches := make(map[string]int)
	matches["min"], matches["max"] = 1, 100 // if no range specified

	match := reRandom.FindStringSubmatch(r.remainder)

	if len(match) > 0 {
		for i, name := range reRandom.SubexpNames() {
			if i == 0 {
				continue
			}
			if integer, err := strconv.Atoi(match[i]); err == nil {
				matches[name] = integer
			}
		}
		if matches["max"] < matches["min"] {
			matches["min"], matches["max"] = matches["max"], matches["min"]
		}
	}

	r.Solution = strconv.Itoa(rand.Intn(matches["max"]+1-matches["min"]) + matches["min"])

	return r
}

func (r *Random) tests() []test {
	tests := []test{}

	solutions := func(choices []string) []Data {
		sol := []Data{}

		for _, c := range choices {
			sol = append(sol,
				Data{
					Type:      RandomType,
					Triggered: true,
					Solution:  c,
				},
			)
		}

		return sol
	}

	for _, c := range []struct {
		q   string
		sol []string
	}{
		{"Random number between 1 and 3", []string{"1", "2", "3"}},
		{"Random number between 5431 and 5434", []string{"5431", "5432", "5433", "5434"}},
		{"Random number between -18 and -21", []string{"-18", "-19", "-20", "-21"}},
	} {
		t := test{
			query:    c.q,
			expected: solutions(c.sol),
		}
		tests = append(tests, t)
	}

	return tests
}

func init() {
	reRandom = regexp.MustCompile(`(?P<min>-?\d+).*?(?P<max>-?\d+)`)
}
