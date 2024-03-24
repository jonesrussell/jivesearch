package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// UserAgentType is an answer Type
const UserAgentType Type = "user agent"

// UserAgent is an instant answer
type UserAgent struct {
	Answer
}

func (u *UserAgent) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *UserAgent) setUserAgent(r *http.Request) Answerer {
	u.Answer.userAgent = r.UserAgent()
	return u
}

func (u *UserAgent) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *UserAgent) setType() Answerer {
	u.Type = UserAgentType
	return u
}

func (u *UserAgent) setRegex() Answerer {
	triggers := []string{
		"user agent", "user agent?",
		"useragent", "useragent?",
		"my user agent", "my user agent?",
		"my useragent", "my useragent?",
		"what's my user agent", "what's my user agent?",
		"what's my useragent", "what's my useragent?",
		"what is my user agent", "what is my user agent?",
		"what is my useragent", "what is my useragent?",
	}

	t := strings.Join(triggers, "|")
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, t)))

	return u
}

func (u *UserAgent) solve(r *http.Request) Answerer {
	u.Solution = u.userAgent
	return u
}

func (u *UserAgent) tests() []test {
	tests := []test{
		{
			query:     "user agent",
			userAgent: "firefox",
			expected: []Data{
				{
					Type:      UserAgentType,
					Triggered: true,
					Solution:  "firefox",
				},
			},
		},
		{
			query:     "useragent?",
			userAgent: "opera",
			expected: []Data{
				{
					Type:      UserAgentType,
					Triggered: true,
					Solution:  "opera",
				},
			},
		},
		{
			query:     "my user agent",
			userAgent: "some random ua",
			expected: []Data{
				{
					Type:      UserAgentType,
					Triggered: true,
					Solution:  "some random ua",
				},
			},
		},
		{
			query:     "what's my user agent?",
			userAgent: "chrome",
			expected: []Data{
				{
					Type:      UserAgentType,
					Triggered: true,
					Solution:  "chrome",
				},
			},
		},
		{
			query:     "what is my useragent?",
			userAgent: "internet explorer",
			expected: []Data{
				{
					Type:      UserAgentType,
					Triggered: true,
					Solution:  "internet explorer",
				},
			},
		},
	}

	return tests
}
