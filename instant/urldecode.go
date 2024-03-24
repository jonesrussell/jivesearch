package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// URLDecodeType is an answer Type
const URLDecodeType Type = "urldecode"

// URLDecode is an instant answer
type URLDecode struct {
	Answer
}

func (u *URLDecode) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *URLDecode) setUserAgent(r *http.Request) Answerer {
	return u
}

func (u *URLDecode) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *URLDecode) setType() Answerer {
	u.Type = URLDecodeType
	return u
}

func (u *URLDecode) setRegex() Answerer {
	triggers := []string{
		"urldecode", "decodeurl", "url decode", "decode url", "urlunescape", "urlunescaper", "unescapeurl", "url unescape", "url unescaper", "unescape url",
		"uridecode", "decodeuri", "uri decode", "decode uri", "uriunescape", "uriunescaper", "unescapeuri", "uri unescape", "uri unescaper", "unescape uri",
	}

	t := strings.Join(triggers, "|")
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	u.regex = append(u.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return u
}

func (u *URLDecode) solve(r *http.Request) Answerer {
	u.Solution, u.Err = url.QueryUnescape(u.remainder)
	return u
}

func (u *URLDecode) tests() []test {
	tests := []test{
		{
			query: "urldecode http%3A%2F%2Fwww.example.com%3Fq%3Dthis%7Cthat",
			expected: []Data{
				{
					Type:      URLDecodeType,
					Triggered: true,
					Solution:  "http://www.example.com?q=this|that",
				},
			},
		},
	}

	return tests
}
