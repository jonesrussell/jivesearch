package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/jonesrussell/jivesearch/instant/shortener"
	"golang.org/x/text/language"
)

// URLShortenerType is an answer Type
const URLShortenerType Type = "url shortener"

// Shortener is an instant answer
type Shortener struct {
	Service shortener.Service
	Answer
}

// ErrInvalidURL indicates an invalid url
var ErrInvalidURL = fmt.Errorf("unable to parse url")

func (s *Shortener) setQuery(r *http.Request, qv string) Answerer {
	s.Answer.setQuery(r, qv)
	return s
}

func (s *Shortener) setUserAgent(r *http.Request) Answerer {
	return s
}

func (s *Shortener) setLanguage(lang language.Tag) Answerer {
	s.language = lang
	return s
}

func (s *Shortener) setType() Answerer {
	s.Type = URLShortenerType
	return s
}

func (s *Shortener) setRegex() Answerer {
	triggers := []string{
		"shorten", "shortener", "short url", "shorten url", "url short", "url shorten", "url shortener",
	}

	t := strings.Join(triggers, "|")
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return s
}

func (s *Shortener) solve(r *http.Request) Answerer {
	u, err := url.Parse(s.remainder)
	if err != nil {
		s.Err = ErrInvalidURL
		return s
	}

	resp, err := s.Service.Shorten(u)
	if err != nil {
		s.Err = err
		return s
	}

	s.Data.Solution = resp
	return s
}

func (s *Shortener) tests() []test {
	u := "https://verylong.com/link"
	original, _ := url.Parse(u)
	shrt, _ := url.Parse("http://shrt.url")

	tests := []test{
		{
			query: fmt.Sprintf("shorten %v", u),
			expected: []Data{
				{
					Type:      URLShortenerType,
					Triggered: true,
					Solution: &shortener.Response{
						Original: original,
						Short:    shrt,
						Provider: "mockShortener",
					},
				},
			},
		},
	}

	return tests
}
