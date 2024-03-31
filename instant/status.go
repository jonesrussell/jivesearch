package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jonesrussell/jivesearch/instant/status"
	"golang.org/x/text/language"
)

// StatusType is an answer Type
const StatusType Type = "status"

// Status is an instant answer
type Status struct {
	Fetcher status.Fetcher
	Answer
}

func (s *Status) setQuery(r *http.Request, qv string) Answerer {
	s.Answer.setQuery(r, qv)
	return s
}

func (s *Status) setUserAgent(r *http.Request) Answerer {
	return s
}

func (s *Status) setLanguage(lang language.Tag) Answerer {
	s.language = lang
	return s
}

func (s *Status) setType() Answerer {
	s.Type = StatusType
	return s
}

func (s *Status) setRegex() Answerer {
	triggers := []string{
		"down", "up", "status", "status of", "is it up", "is it down", "isitup", "isitdown", "working",
	}

	t := strings.Join(triggers, "|")
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return s
}

func (s *Status) solve(r *http.Request) Answerer {
	d := status.FixDomain(s.remainder)

	resp, err := s.Fetcher.Fetch(d)
	if err != nil {
		s.Err = err
		return s
	}

	s.Data.Solution = resp
	return s
}

func (s *Status) tests() []test {
	tests := []test{
		{
			query: "is it up example.com",
			expected: []Data{
				{
					Type:      StatusType,
					Triggered: true,
					Solution: &status.Response{
						Domain:   "example.com",
						Port:     80,
						Status:   1,
						IP:       "1.2.3.4",
						Code:     200,
						Time:     1.2,
						Provider: status.IsItUpProvider,
					},
				},
			},
		},
		{
			query: "isitdown something",
			expected: []Data{
				{
					Type:      StatusType,
					Triggered: true,
					Solution: &status.Response{
						Domain:   "something.com",
						Port:     80,
						Status:   1,
						IP:       "1.2.3.5",
						Code:     301,
						Time:     .23,
						Provider: status.IsItUpProvider,
					},
				},
			},
		},
	}

	return tests
}
