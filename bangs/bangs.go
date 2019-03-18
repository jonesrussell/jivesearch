// Package bangs provides functionality to query other websites
package bangs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

// Bangs holds a map of !bangs
type Bangs struct {
	Bangs []Bang `mapstructure:"bang"`
	Suggester
}

// Bang holds a single !bang
type Bang struct {
	Name      string            `json:"name"`
	FavIcon   string            `json:"favicon"`
	Triggers  []string          `json:"triggers"`
	Regions   map[string]string `json:"regions"`
	Functions []string          `json:"-"`
	Funcs     []fn              `json:"-"`
}

// Suggester is a !bangs suggester/autocomplete
type Suggester interface {
	IndexExists() (bool, error)
	DeleteIndex() error
	Setup([]Bang) error
	SuggestResults(term string, size int) (Results, error)
}

// Results are the results of an autocomplete query
type Results struct { // remember top-level arrays = no-no in javascript/json
	Suggestions []Suggestion `json:"suggestions"`
}

// Suggestion is an individual !bang autocomplete suggestion
type Suggestion struct {
	Trigger string `json:"trigger"`
	Name    string `json:"name"`
	FavIcon string `json:"favicon"`
}

const def = "default"

// Provider is a configuration provider
type Provider interface {
	ReadInConfig() error
	Unmarshal(interface{}, ...viper.DecoderConfigOption) error
}

// New creates Bangs from a config file
func New(cfg Provider) (*Bangs, error) {
	var b = &Bangs{}

	if err := cfg.ReadInConfig(); err != nil {
		return nil, err
	}

	err := cfg.Unmarshal(&b, viperStructTag())
	return b, err
}

func viperStructTag() viper.DecoderConfigOption {
	return func(c *mapstructure.DecoderConfig) {
	}
}

// Suggest is an autocomplete for !bangs
func (b *Bangs) Suggest(term string, size int) (Results, error) {
	res, err := b.Suggester.SuggestResults(term, size)
	if err != nil {
		return res, err
	}

	// fill in the rest of the suggestion
	for i, s := range res.Suggestions {
		for _, bng := range b.Bangs {
			for _, trigger := range bng.Triggers {
				if trigger == s.Trigger {
					s.Name = bng.Name
					s.FavIcon = bng.FavIcon
					res.Suggestions[i] = s
				}
			}
		}
	}

	return res, err
}

// Detect lets us know if we have a !bang match.
func (b *Bangs) Detect(q string, region language.Region, l language.Tag) (Bang, string, bool) {
	fields := strings.Fields(q)

	for i, field := range fields {
		if field == "!" || (!strings.HasPrefix(field, "!") && !strings.HasSuffix(field, "!")) {
			continue
		}

		k := strings.ToLower(strings.Trim(field, "!"))
		for _, bng := range b.Bangs {
			if triggered := trigger(k, bng.Triggers); !triggered {
				continue
			}

			remainder := strings.Join(append(fields[:i], fields[i+1:]...), " ")

			for _, f := range bng.Funcs {
				remainder = f(remainder)
			}

			for _, reg := range []string{strings.ToLower(region.String()), def} { // use default region if no region specified
				if u, ok := bng.Regions[reg]; ok {
					u = strings.Replace(u, "{{{term}}}", url.QueryEscape(remainder), -1)
					return bng, strings.Replace(u, "{{{lang}}}", l.String(), -1), true
				}
			}
		}
	}
	return Bang{}, "", false
}

type fn func(string) string

// Returns the canonical version of a Wikipedia title.
// "bob maRLey" -> "Bob_Marley"
// Some regional queries will be incorrect: https://es.wikipedia.org/wiki/De_la_Tierra_a_la_Luna
var wikipediaCanonical = func(q string) string {
	return strings.Replace(strings.Title(strings.ToLower(q)), " ", "_", -1)
}

func trigger(k string, triggers []string) bool {
	for _, trigger := range triggers {
		if k == trigger {
			return true
		}
	}
	return false
}

// CreateFunctions creates []Funcs from []Functions.
// Is a workaround since I couldn't find a way to map a function type in a config file.
func (b *Bangs) CreateFunctions() error {
	for i, bng := range b.Bangs {
		for _, f := range bng.Functions {
			var ff fn

			switch f {
			case "wikipediaCanonical":
				ff = wikipediaCanonical
			default:
				return fmt.Errorf("unknown function string %v", f)
			}
			b.Bangs[i].Funcs = append(b.Bangs[i].Funcs, ff)
		}
	}
	return nil
}
