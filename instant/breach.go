package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/breach"
	"golang.org/x/text/language"
)

// BreachType is an answer Type
const BreachType Type = "breach"

// Breach is an instant answer
type Breach struct {
	Fetcher breach.Fetcher
	Answer
}

func (b *Breach) setQuery(r *http.Request, qv string) Answerer {
	b.Answer.setQuery(r, qv)
	return b
}

func (b *Breach) setUserAgent(r *http.Request) Answerer {
	return b
}

func (b *Breach) setLanguage(lang language.Tag) Answerer {
	b.language = lang
	return b
}

func (b *Breach) setType() Answerer {
	b.Type = BreachType
	return b
}

func (b *Breach) setRegex() Answerer {
	triggers := []string{
		"breach", "pwned", "have i been pwned",
	}

	t := strings.Join(triggers, "|")
	b.regex = append(b.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	b.regex = append(b.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return b
}

func (b *Breach) solve(r *http.Request) Answerer {
	resp, err := b.Fetcher.Fetch(b.remainder)
	if err != nil {
		b.Err = err
		return b
	}

	resp.Sort()
	b.Data.Solution = resp
	return b
}

func (b *Breach) tests() []test {
	accnt := "test@example.com"

	tests := []test{
		{
			query: fmt.Sprintf("pwned %v", accnt),
			expected: []Data{
				{
					Type:      BreachType,
					Triggered: true,
					Solution: &breach.Response{
						Account: "test@example.com",
						Breaches: []breach.Breach{
							{
								Name:        "000webhost",
								Domain:      "000webhost.com",
								Date:        time.Date(2015, 3, 1, 0, 0, 0, 0, time.UTC),
								Count:       14936670,
								Description: "Some description here.",
								Items:       []string{"Email addresses", "IP addresses", "Names", "Passwords"},
							},
							{
								Name:        "8tracks",
								Domain:      "8tracks.com",
								Date:        time.Date(2017, 6, 27, 0, 0, 0, 0, time.UTC),
								Count:       7990619,
								Description: "Another description here.",
								Items:       []string{"Email addresses", "Passwords"},
							},
						},
						Provider: breach.HaveIBeenPwnedProvider,
					},
				},
			},
		},
	}

	return tests
}
