package instant

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/jivesearch/jivesearch/instant/whois"
	"golang.org/x/text/language"
)

// WHOISType is an answer Type
const WHOISType Type = "whois"

// WHOIS is an instant answer
type WHOIS struct {
	Answer
	whois.Fetcher
}

func (w *WHOIS) setQuery(r *http.Request, qv string) Answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *WHOIS) setUserAgent(r *http.Request) Answerer {
	return w
}

func (w *WHOIS) setLanguage(lang language.Tag) Answerer {
	w.language = lang
	return w
}

func (w *WHOIS) setType() Answerer {
	w.Type = WHOISType
	return w
}

func (w *WHOIS) setRegex() Answerer {
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>whois) (?P<remainder>.*)$`)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>whois)$`)))

	return w
}

func (w *WHOIS) solve(r *http.Request) Answerer {
	resp, err := w.Fetch(w.remainder)
	if err != nil {
		w.Err = err
		return w
	}

	if resp.Error != "" {
		w.Err = fmt.Errorf(resp.Error)
		return w
	}

	w.Data.Solution = resp
	return w
}

func (w *WHOIS) tests() []test {
	tests := []test{
		{
			query: "whois google.com",
			expected: []Data{
				{
					Type:      WHOISType,
					Triggered: true,
					Solution: &whois.Response{
						Domain:   "google.com",
						DomainID: "2138514_DOMAIN_COM-VRSN",
						Status:   "registered",
						Nameservers: []whois.NameServer{
							{Name: "ns3.google.com"}, {Name: "ns4.google.com"}, {Name: "ns1.google.com"}, {Name: "ns2.google.com"},
						},
						Available:  false,
						Registered: true,
					},
				},
			},
		},
	}

	return tests
}
