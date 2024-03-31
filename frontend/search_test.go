package frontend

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jonesrussell/jivesearch/bangs"
	"github.com/jonesrussell/jivesearch/instant"
	"github.com/jonesrussell/jivesearch/instant/breach"
	"github.com/jonesrussell/jivesearch/instant/stackoverflow"
	"github.com/jonesrussell/jivesearch/instant/wikipedia"
	"github.com/jonesrussell/jivesearch/search"
	"github.com/jonesrussell/jivesearch/search/document"
	img "github.com/jonesrussell/jivesearch/search/image"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

func bangsFromConfig() (*bangs.Bangs, error) {
	vb := viper.New()
	vb.SetConfigType("toml")
	vb.SetConfigName("bangs")
	vb.AddConfigPath("../bangs")
	return bangs.New(vb)
}

func getDefaultBangs(bngs *bangs.Bangs) ([]DefaultBang, error) {
	var m = map[string]bangs.Bang{}

	for _, b := range bngs.Bangs {
		if b.Name == "Google" {
			m["g"] = b
		}
		if b.Name == "Bing" {
			m["b"] = b
		}
		if b.Name == "Amazon" {
			m["a"] = b
		}
		if b.Name == "YouTube" {
			m["yt"] = b
		}
	}

	db := []DefaultBang{
		{"g", m["g"]}, {"b", m["b"]}, {"a", m["a"]}, {"yt", m["yt"]},
	}

	return db, nil
}

func TestDefaultBangs(t *testing.T) {
	var bing = DefaultBang{}
	var so = DefaultBang{}
	bngs, err := bangsFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	for _, b := range bngs.Bangs {
		if b.Name == "Bing" {
			bing = DefaultBang{"b", b}
		}
		if b.Name == "Stack Overflow" {
			so = DefaultBang{"so", b}
		}
	}

	db, err := getDefaultBangs(bngs)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name string
		args string
		want []DefaultBang
	}{
		{
			"default", "", db,
		},
		{
			"custom", "b,so", []DefaultBang{bing, so},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Bangs: bngs,
			}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("b", c.args)
			req.URL.RawQuery = q.Encode()

			got := f.defaultBangs(req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectLanguage(t *testing.T) {
	for _, c := range []struct {
		name           string
		acceptLanguage string
		l              string
		want           []language.Tag
	}{
		{
			"blank", "", "", []language.Tag{},
		},
		{
			"basic", "", "en", []language.Tag{language.English},
		},
		{
			"french", "", "fr", []language.Tag{language.French},
		},
		{
			"Accept-Language header",
			"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7",
			"",
			[]language.Tag{
				language.MustParse("fr-CH"),
				language.French,
				language.English,
				language.German,
			},
		},
		{
			"param overrides Accept-Language header",
			"fr-CH, fr;q=0.9, en;q=0.8, de;q=0.7",
			"hr",
			[]language.Tag{
				language.Croatian,
				language.MustParse("fr-CH"),
				language.French,
				language.English,
				language.German,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			req.Header.Set("Accept-Language", c.acceptLanguage)

			q := req.URL.Query()
			q.Add("l", c.l)

			req.URL.RawQuery = q.Encode()

			got := f.detectLanguage(req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectRegion(t *testing.T) {
	for _, c := range []struct {
		name string
		lang language.Tag
		r    string
		want language.Region
	}{
		{
			"empty", language.Tag{}, "", language.MustParseRegion("US").Canonicalize(),
		},
		{
			"basic", language.Tag{}, "us", language.MustParseRegion("US").Canonicalize(),
		},
		{
			"region from language", language.BrazilianPortuguese, "", language.MustParseRegion("BR").Canonicalize(),
		},
		{
			"param overrides language's region", language.CanadianFrench, "gb", language.MustParseRegion("GB").Canonicalize(),
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("r", c.r)

			req.URL.RawQuery = q.Encode()

			got := f.detectRegion(c.lang, req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestSearchHandler(t *testing.T) {
	bngs, err := bangsFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	db, err := getDefaultBangs(bngs)
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		name     string
		language string
		query    string
		output   string
		t        string
		safe     string
		filter   string
		theme    string
		want     *response
	}{
		{
			"empty", "en", "", "", "", "", "", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Brand: Brand{},
					Context: &Context{
						F:    search.Moderate,
						Safe: true,
					},
				},
			},
		},
		{
			"basic", "en", " some query ", "", "", "f", "off", "night",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Brand: Brand{},
					Context: &Context{
						Q:            "some query",
						L:            "en",
						lang:         language.MustParse("en"),
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
						F:            search.Off,
						Safe:         false,
						Theme:        "night",
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"not cached", "en", "not cached", "", "", "", "strict", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Brand: Brand{},
					Context: &Context{
						Q:            "not cached",
						L:            "en",
						lang:         language.MustParse("en"),
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
						F:            search.Strict,
						Safe:         true,
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"json", "en", " some query", "json", "", "", "", "",
			&response{
				status:   http.StatusOK,
				template: "json",
				data: data{
					Brand: Brand{},
					Context: &Context{
						Q:            "some query",
						L:            "en",
						lang:         language.MustParse("en"),
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
						F:            search.Moderate,
						Safe:         true,
					},
					Results: Results{
						Instant: mockInstantAnswer,
						Search:  mockSearchResults,
					},
				},
			},
		},
		{
			"!bang", "", "!g something", "", "", "", "", "",
			&response{
				status:   http.StatusFound,
				redirect: "https://encrypted.google.com/search?hl=en&q=something",
			},
		},
		{
			"images", "en", "some query", "", "images", "", "", "",
			&response{
				status:   http.StatusOK,
				template: "search",
				data: data{
					Brand: Brand{},
					Context: &Context{
						Q:            "some query",
						L:            "en",
						lang:         language.MustParse("en"),
						DefaultBangs: db,
						Preferred:    []language.Tag{language.MustParse("en")},
						Region:       language.MustParseRegion("US"),
						Number:       25,
						Page:         1,
						F:            search.Moderate,
						Safe:         true,
						T:            "images",
					},
					Results: Results{
						Images: mockImageResults,
						Search: &search.Results{},
					},
				},
			},
		},
		{
			"first result", "", "! first result", "", "", "", "", "",
			&response{
				status:   http.StatusFound,
				redirect: "https://example.com",
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			var matcher = language.NewMatcher(
				[]language.Tag{
					language.English,
					language.French,
				},
			)

			f := &Frontend{
				Brand: Brand{},
				Document: Document{
					Matcher: matcher,
				},
				Bangs: bngs,
				Instant: &instant.Instant{
					WikipediaFetcher:     &mockWikipediaFetcher{},
					StackOverflowFetcher: &mockStackOverflowFetcher{},
				},
				Suggest: &mockSuggester{},
				Search:  &mockSearch{},
				Wikipedia: Wikipedia{
					Matcher: matcher,
				},
			}

			f.Images.Client = &http.Client{}
			f.Images.Fetcher = &mockImages{}
			f.Cache.Cacher = &mockCacher{}
			f.Cache.Instant = 10 * time.Second
			f.Cache.Search = 10 * time.Second

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("q", c.query)
			q.Add("l", c.language)
			q.Add("o", c.output)
			q.Add("t", c.t)
			q.Add("safe", c.safe)
			q.Add("f", c.filter)
			q.Add("theme", c.theme)
			req.URL.RawQuery = q.Encode()

			got := f.searchHandler(httptest.NewRecorder(), req)

			if gotErr, wantErr := got.err, c.want.err; gotErr != nil || wantErr != nil {
				if (gotErr != nil && wantErr == nil) || (gotErr == nil && wantErr != nil) || gotErr.Error() != wantErr.Error() {
					t.Fatalf("got error %v; want error %v", gotErr, wantErr)
				}
			}
		})
	}
}

type mockSearch struct{}

func (s *mockSearch) Fetch(q string, f search.Filter, lang language.Tag, region language.Region, page int, number int) (*search.Results, error) {
	return mockSearchResults, nil
}

type mockImages struct{}

func (i *mockImages) Fetch(q string, safe bool, number int, offset int) (*img.Results, error) {
	return mockImageResults, nil
}

type mockCacher struct{}

func (c *mockCacher) Get(key string) (interface{}, error) {
	var v interface{}

	switch key {
	case "::instant::en::US::/answer?q=2%2B2":
		v = mockCalculatorInstantAnswer
	case "::instant::en::US::/answer?q=january+birthstone":
		v = mockBirthstoneInstantAnswer
	case "::instant::en::US::/?l=en&o=json&q=+some+query", "::instant::en::US::/?l=en&o=&q=+some+query+":
		v = mockInstantAnswer
	case "::search::en::US::/?l=en&o=json&q=+some+query", "::search::en::US::/?l=en&o=&q=+some+query+":
		v = mockSearchResults
	default:
		return nil, nil
	}

	j, err := json.Marshal(v)
	if err != nil {
		return []int8{}, err
	}

	return j, nil
}

func (c *mockCacher) Put(key string, value interface{}, ttl time.Duration) error {
	return nil
}

// mock BreachFetcher
type mockBreachFetcher struct{}

func (m *mockBreachFetcher) Fetch(account string) (*breach.Response, error) {
	return &breach.Response{}, nil
}

// mock Stack Overflow Fetcher
type mockStackOverflowFetcher struct{}

func (s *mockStackOverflowFetcher) Fetch(query string, tags []string) (stackoverflow.Response, error) {
	return stackoverflow.Response{}, nil
}

// mock Wikipedia Fetcher
type mockWikipediaFetcher struct{}

func (mf *mockWikipediaFetcher) Fetch(query string, lang language.Tag) ([]*wikipedia.Item, error) {
	return []*wikipedia.Item{{}}, nil
}

func (mf *mockWikipediaFetcher) Setup() error {
	return nil
}

var mockBirthstoneInstantAnswer = instant.Data{
	Type:      "birthstone",
	Triggered: true,
	Solution:  "Garnet",
	Err:       nil,
}

var mockCalculatorInstantAnswer = instant.Data{
	Type:      "calculator",
	Triggered: true,
	Solution:  "4",
	Err:       nil,
}

var mockInstantAnswer = instant.Data{
	Type:      "wikipedia",
	Triggered: true,
	Solution:  []*wikipedia.Item{{}},
	Err:       nil,
}

var mockSearchResults = &search.Results{
	Count:      int64(25),
	Page:       "1",
	Previous:   "",
	Next:       "2",
	Last:       "72",
	Pagination: []string{"1"},
	Documents: []*document.Document{
		{ID: "https://example.com"},
		{ID: "https://examples.com"},
	},
}

var mockImageResults = &img.Results{
	Count:      int64(25),
	Page:       "1",
	Previous:   "",
	Next:       "2",
	Last:       "72",
	Pagination: []string{"1"},
	Images:     []*img.Image{},
}
