package instant

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	disc "github.com/jivesearch/jivesearch/instant/discography"

	"golang.org/x/text/language"
)

// DiscographyType is an answer Type
const DiscographyType Type = "discography"

// Discography is an instant answer
type Discography struct {
	disc.Fetcher
	Answer
}

func (d *Discography) setQuery(req *http.Request, q string) Answerer {
	d.Answer.setQuery(req, q)
	return d
}

func (d *Discography) setUserAgent(req *http.Request) Answerer {
	return d
}

func (d *Discography) setLanguage(lang language.Tag) Answerer {
	d.language = lang
	return d
}

func (d *Discography) setType() Answerer {
	d.Type = DiscographyType
	return d
}

func (d *Discography) setRegex() Answerer {
	triggers := []string{
		"discography", "albums",
	}

	t := strings.Join(triggers, "|")
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return d
}

func (d *Discography) solve(r *http.Request) Answerer {
	albums, err := d.Fetch(d.remainder)
	if err != nil {
		d.Err = err
		return d
	}

	d.Data.Solution = albums
	return d
}

func (d *Discography) tests() []test {
	tests := []test{
		{
			query: "jimi hendrix discography",
			expected: []Data{
				{
					Type:      DiscographyType,
					Triggered: true,
					Solution: []disc.Album{
						{
							Name:      "Are You Experienced",
							Published: time.Date(1970, 9, 18, 0, 0, 0, 0, time.UTC),
							Image: disc.Image{
								URL: discURL,
							},
						},
					},
				},
			},
		},
	}
	return tests
}

var discURL, _ = url.Parse("http://coverartarchive.org/release/1/2-250..jpg")
