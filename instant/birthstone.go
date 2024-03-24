package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// BirthStoneType is an answer Type
const BirthStoneType Type = "birthstone"

// BirthStone is an instant answer
type BirthStone struct {
	Answer
}

func (b *BirthStone) setQuery(r *http.Request, qv string) Answerer {
	b.Answer.setQuery(r, qv)
	return b
}

func (b *BirthStone) setUserAgent(r *http.Request) Answerer {
	return b
}

func (b *BirthStone) setLanguage(lang language.Tag) Answerer {
	b.language = lang
	return b
}

func (b *BirthStone) setType() Answerer {
	b.Type = BirthStoneType
	return b
}

func (b *BirthStone) setRegex() Answerer {
	triggers := []string{
		"birthstones",
		"birth stones",
		"birthstone",
		"birth stone",
	}

	t := strings.Join(triggers, "|")
	b.regex = append(b.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	b.regex = append(b.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))

	return b
}

func (b *BirthStone) solve(r *http.Request) Answerer {
	switch b.remainder {
	case "january":
		b.Solution = "Garnet"
	case "february":
		b.Solution = "Amethyst"
	case "march":
		b.Solution = "Aquamarine, Bloodstone"
	case "april":
		b.Solution = "Diamond"
	case "may":
		b.Solution = "Emerald"
	case "june":
		b.Solution = "Pearl, Moonstone, Alexandrite"
	case "july":
		b.Solution = "Ruby"
	case "august":
		b.Solution = "Peridot, Spinel"
	case "september":
		b.Solution = "Sapphire"
	case "october":
		b.Solution = "Opal, Tourmaline"
	case "november":
		b.Solution = "Topaz, Citrine"
	case "december":
		b.Solution = "Turquoise, Zircon, Tanzanite"
	}

	return b
}

func (b *BirthStone) tests() []test {
	tests := []test{
		{
			query: "January birthstone",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Garnet",
				},
			},
		},
		{
			query: "birthstone february",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Amethyst",
				},
			},
		},
		{
			query: "march birth stone",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Aquamarine, Bloodstone",
				},
			},
		},
		{
			query: "birth stone April",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Diamond",
				},
			},
		},
		{
			query: "birth stones may",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Emerald",
				},
			},
		},
		{
			query: "birthstones June",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Pearl, Moonstone, Alexandrite",
				},
			},
		},
		{
			query: "July Birth Stones",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Ruby",
				},
			},
		},
		{
			query: "birthstones August",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Peridot, Spinel",
				},
			},
		},
		{
			query: "september birthstones",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Sapphire",
				},
			},
		},
		{
			query: "October birthstone",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Opal, Tourmaline",
				},
			},
		},
		{
			query: "birthstone November",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Topaz, Citrine",
				},
			},
		},
		{
			query: "December birthstone",
			expected: []Data{
				{
					Type:      BirthStoneType,
					Triggered: true,
					Solution:  "Turquoise, Zircon, Tanzanite",
				},
			},
		},
	}

	return tests
}
