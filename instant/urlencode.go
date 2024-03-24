package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// URLEncodeType is an answer Type
const URLEncodeType Type = "urlencode"

// URLEncode is an instant answer
type URLEncode struct {
	Answer
}

func (u *URLEncode) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *URLEncode) setUserAgent(r *http.Request) Answerer {
	return u
}

func (u *URLEncode) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *URLEncode) setType() Answerer {
	u.Type = URLEncodeType
	return u
}

func (u *URLEncode) setRegex() Answerer {
	triggers := []string{
		"urlencode", "encodeurl", "url encode", "encode url", "urlescape", "urlescaper", "escapeurl", "url escape", "url escaper", "escape url",
		"uriencode", "encodeuri", "uri encode", "encode uri", "uriescape", "uriescaper", "escapeuri", "uri escape", "uri escaper", "escape uri",
	}

	t := strings.Join(triggers, "|")
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return u
}

func (u *URLEncode) solve(r *http.Request) Answerer {
	u.Solution = url.QueryEscape(u.remainder)
	return u
}

func (u *URLEncode) tests() []test {
	tests := []test{
		{
			query: "urlencode http://www.example.com?q=this|that",
			expected: []Data{
				{
					Type:      URLEncodeType,
					Triggered: true,
					Solution:  "http%3A%2F%2Fwww.example.com%3Fq%3Dthis%7Cthat",
				},
			},
		},
	}

	return tests
}
