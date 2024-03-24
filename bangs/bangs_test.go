package bangs

import (
	"reflect"
	"testing"

	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

func fromConfig() (*Bangs, error) {
	vb := viper.New()
	vb.SetConfigType("toml")
	vb.SetConfigName("bangs")
	vb.AddConfigPath("../bangs")
	return New(vb)
}

// TestDefault tests that each !bang has a default location
func TestDefault(t *testing.T) {
	b, err := fromConfig()
	if err != nil {
		t.Fatal(err)
	}

	for _, bng := range b.Bangs {
		if _, ok := bng.Regions[def]; !ok {
			t.Fatalf("%q bang needs a default region", bng.Name)
		}
	}
}

// TestFavIcon tests that each !bang has a favicon
func TestFavIcon(t *testing.T) {
	b, err := fromConfig()
	if err != nil {
		t.Fatal(err)
	}

	for _, bng := range b.Bangs {
		if bng.FavIcon == "" {
			t.Fatalf("%q bang needs a favicon", bng.Name)
		}
	}

}

func TestDuplicateTriggers(t *testing.T) {
	seen := make(map[string]bool)

	b, err := fromConfig()
	if err != nil {
		t.Fatal(err)
	}
	for _, bng := range b.Bangs {
		for _, trig := range bng.Triggers {
			if _, ok := seen[trig]; ok {
				t.Fatalf("duplicate trigger found %q", trig)
			}
			seen[trig] = true
		}
	}
}

func TestSuggest(t *testing.T) {
	type args struct {
		term string
		size int
	}

	for _, c := range []struct {
		name string
		args
		want Results
	}{
		{
			"basic",
			args{"g", 10},
			Results{Suggestions: []Suggestion{
				{"g", "Google", "https://www.google.com/favicon.ico"},
				{"gfr", "Google France", "https://www.google.com/favicon.ico"},
				{"gh", "GitHub", "https://github.com/favicon.ico"},
				{"gi", "Google Images", "https://www.google.com/favicon.ico"},
			}},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			b, err := fromConfig()
			if err != nil {
				t.Fatal(err)
			}
			b.Suggester = &mockSuggester{}
			got, err := b.Suggest(c.args.term, c.args.size)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestDetect(t *testing.T) {
	type data struct {
		b   Bang
		loc string
		ok  bool
	}

	for _, c := range []struct {
		q    string
		r    string
		l    language.Tag
		want data
	}{
		{
			q: "!g bob", r: "US", l: language.French,
			want: data{
				b: Bang{
					Name:     "Google",
					FavIcon:  "https://www.google.com/favicon.ico",
					Triggers: []string{"g", "google"},
					Regions: map[string]string{
						"ru":      "https://www.google.ru/search?hl={{{lang}}}&q={{{term}}}",
						"default": "https://encrypted.google.com/search?hl={{{lang}}}&q={{{term}}}",
						"ca":      "https://www.google.ca/search?q={{{term}}}",
						"fr":      "https://www.google.fr/search?hl={{{lang}}}&q={{{term}}}",
					},
				},
				loc: "https://encrypted.google.com/search?hl=fr&q=bob",
				ok:  true,
			},
		},
		{
			q: "!g at&t", r: "US", l: language.English,
			want: data{
				b: Bang{
					Name:     "Google",
					FavIcon:  "https://www.google.com/favicon.ico",
					Triggers: []string{"g", "google"},
					Regions: map[string]string{
						"ru":      "https://www.google.ru/search?hl={{{lang}}}&q={{{term}}}",
						"default": "https://encrypted.google.com/search?hl={{{lang}}}&q={{{term}}}",
						"ca":      "https://www.google.ca/search?q={{{term}}}",
						"fr":      "https://www.google.fr/search?hl={{{lang}}}&q={{{term}}}",
					},
				},
				loc: "https://encrypted.google.com/search?hl=en&q=at%26t",
				ok:  true,
			},
		},
		{
			q: "!g bob french", r: "fr", l: language.English,
			want: data{
				b: Bang{
					Name:     "Google",
					FavIcon:  "https://www.google.com/favicon.ico",
					Triggers: []string{"g", "google"},
					Regions: map[string]string{
						"ru":      "https://www.google.ru/search?hl={{{lang}}}&q={{{term}}}",
						"default": "https://encrypted.google.com/search?hl={{{lang}}}&q={{{term}}}",
						"ca":      "https://www.google.ca/search?q={{{term}}}",
						"fr":      "https://www.google.fr/search?hl={{{lang}}}&q={{{term}}}",
					},
				},
				loc: "https://www.google.fr/search?hl=en&q=bob+french",
				ok:  true,
			},
		},
		{
			q: "!gfr something french", r: "fr", l: language.English,
			want: data{
				b: Bang{
					Name:     "Google France",
					FavIcon:  "https://www.google.com/favicon.ico",
					Triggers: []string{"gfr", "googlefr"},
					Regions: map[string]string{
						"default": "https://www.google.fr/search?hl={{{lang}}}&q={{{term}}}",
					},
				},
				loc: "https://www.google.fr/search?hl=en&q=something+french",
				ok:  true,
			},
		},
		{
			q: "!W bob maRLey", r: "US", l: language.French,
			want: data{
				b: Bang{
					Name:     "Wikipedia",
					FavIcon:  "https://en.wikipedia.org/favicon.ico",
					Triggers: []string{"w", "wikipedia", "wiki", "encyclopedia", "wen"},
					Regions: map[string]string{
						"default": "https://en.wikipedia.org/wiki/{{{term}}}",
						"es":      "https://es.wikipedia.org/wiki/{{{term}}}",
						"de":      "https://de.wikipedia.org/wiki/{{{term}}}",
						"fr":      "https://fr.wikipedia.org/wiki/{{{term}}}",
					},
					Functions: []string{"wikipediaCanonical"},
					Funcs:     []fn{wikipediaCanonical},
				},
				loc: "https://en.wikipedia.org/wiki/Bob_Marley",
				ok:  true,
			},
		},
		{
			q: "nonexistent! some query", r: "US", l: language.French,
			want: data{
				b:   Bang{},
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang", r: "US", l: language.English,
			want: data{
				b:   Bang{},
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang g", r: "US", l: language.English,
			want: data{
				b:   Bang{},
				loc: "",
				ok:  false,
			},
		},
		{
			q: "this is not a bang google", r: "US", l: language.English,
			want: data{
				b:   Bang{},
				loc: "",
				ok:  false,
			},
		},
	} {
		t.Run(c.q, func(t *testing.T) {
			b, err := fromConfig()
			if err != nil {
				t.Fatal(err)
			}

			if err := b.CreateFunctions(); err != nil {
				t.Fatal(err)
			}

			r := language.MustParseRegion(c.r)

			var got = data{}
			got.b, got.loc, got.ok = b.Detect(c.q, r, c.l)

			if !reflect.DeepEqual(got, c.want) {
				if got.b.Name != c.want.b.Name || got.loc != c.want.loc || got.ok != c.want.ok {
					t.Fatalf("got %+v; want %+v", got, c.want)
				}
			}
		})
	}
}

type mockSuggester struct{}

func (m *mockSuggester) SuggestResults(term string, size int) (Results, error) {
	res := Results{Suggestions: []Suggestion{
		{Trigger: "g"}, {Trigger: "gfr"}, {Trigger: "gh"}, {Trigger: "gi"},
	}}

	return res, nil
}

func (m *mockSuggester) IndexExists() (bool, error) {
	return true, nil
}

func (m *mockSuggester) DeleteIndex() error {
	return nil
}

func (m *mockSuggester) Setup(bangs []Bang) error {
	return nil
}
