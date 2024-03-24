package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/jivesearch/jivesearch/instant/congress"
	"golang.org/x/text/language"
)

// CongressType is an answer Type
const CongressType Type = "congress"

// Congress is an instant answer
type Congress struct {
	Fetcher congress.Fetcher
	Answer
}

func (c *Congress) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Congress) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *Congress) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *Congress) setType() Answerer {
	c.Type = CongressType
	return c
}

func (c *Congress) setRegex() Answerer {
	senate := []string{
		"senate", "senator", "senators",
	}

	t := strings.Join(senate, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<senate>%s) (?P<state>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<state>.*) (?P<senate>%s)$`, t)))

	members := []string{
		"member", "members", "house members", "congress",
	}

	t = strings.Join(members, "|")
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<members>%s) (?P<state>.*)$`, t)))
	c.regex = append(c.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<state>.*) (?P<members>%s)$`, t)))

	return c
}

func (c *Congress) solve(r *http.Request) Answerer {
	state, ok := c.remainderM["state"]
	if !ok {
		return c
	}

	// validate state
	loc := congress.ValidateState(state)
	if loc == nil {
		c.Err = congress.ErrInvalidState
		return c
	}

	if _, ok := c.remainderM["members"]; ok {
		resp, err := c.Fetcher.FetchMembers(loc)
		if err != nil {
			c.Err = err
			return c
		}
		c.Data.Solution = resp
		return c
	}

	// Senate
	resp, err := c.Fetcher.FetchSenators(loc)
	if err != nil {
		c.Err = err
		return c
	}

	c.Data.Solution = resp
	return c
}

func (c *Congress) tests() []test {
	tests := []test{
		{
			query: "utah members",
			expected: []Data{
				{
					Type:      CongressType,
					Triggered: true,
					Solution: &congress.Response{
						Location: &congress.Location{
							Short: "UT",
							State: "Utah",
						},
						Role: congress.House,
						Members: []congress.Member{
							{
								Name:         "Rob Bishop",
								District:     1,
								Gender:       "M",
								Party:        "R",
								Twitter:      "RepRobBishop",
								Facebook:     "RepRobBishop",
								NextElection: 2018,
							},
							{
								Name:         "Chris Stewart",
								District:     2,
								Gender:       "M",
								Party:        "R",
								Twitter:      "RepChrisStewart",
								Facebook:     "RepChrisStewart",
								NextElection: 2018,
							},
							{
								Name:         "John Curtis",
								District:     3,
								Gender:       "M",
								Party:        "R",
								Twitter:      "RepJohnCurtis",
								Facebook:     "",
								NextElection: 2018,
							},
							{
								Name:         "Mia Love",
								District:     4,
								Gender:       "F",
								Party:        "R",
								Twitter:      "repmialove",
								Facebook:     "",
								NextElection: 2018,
							},
						},
						Provider: congress.ProPublicaProvider,
					},
				},
			},
		},
		{
			query: "utah senators",
			expected: []Data{
				{
					Type:      CongressType,
					Triggered: true,
					Solution: &congress.Response{
						Location: &congress.Location{
							Short: "UT",
							State: "Utah",
						},
						Role: congress.Senators,
						Members: []congress.Member{
							{
								Name:         "Orrin G. Hatch",
								Gender:       "M",
								Party:        "R",
								Twitter:      "SenOrrinHatch",
								Facebook:     "senatororrinhatch",
								NextElection: 2018,
							},
							{
								Name:         "Mike Lee",
								Gender:       "M",
								Party:        "R",
								Twitter:      "SenMikeLee",
								Facebook:     "senatormikelee",
								NextElection: 2022,
							},
						},
						Provider: congress.ProPublicaProvider,
					},
				},
			},
		},
	}

	return tests
}
