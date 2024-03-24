package frontend

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/suggest"
)

func TestMiddleware(t *testing.T) {
	type want struct {
		status int
		body   string
	}

	for _, c := range []struct {
		name  string
		tmpl  string
		ct    string
		cl    string
		sniff string
		want
	}{
		{"json", "json", "application/json", "28", "",
			want{http.StatusOK, "{\"response\":\"hello world!\"}\n"},
		},
		{"jsonp", "jsonp", "application/javascript", "48", "",
			want{http.StatusOK, "jivesearchcallback({\"response\":\"hello world!\"}\n)"},
		},
		{"wrong template", "", "text/plain; charset=utf-8", "22", "nosniff",
			want{http.StatusInternalServerError, "Internal Server Error\n"},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			want := make(http.Header)
			want["Content-Length"] = []string{c.cl}
			want["Content-Type"] = []string{c.ct}

			if c.sniff != "" {
				want["X-Content-Type-Options"] = []string{c.sniff}
			}

			f := &Frontend{}
			ParseTemplates()

			fn := func(w http.ResponseWriter, r *http.Request) *response {
				return &response{
					status:   200,
					template: c.tmpl,
					data:     map[string]string{"response": "hello world!"},
					err:      nil,
				}
			}

			ts := httptest.NewServer(f.middleware(appHandler(fn)))
			defer ts.Close()

			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != c.want.status {
				t.Fatalf("got %d; want %d", resp.StatusCode, c.want.status)
			}

			h := resp.Header
			delete(h, "Date") // is there a way to mock this instead???

			if !reflect.DeepEqual(h, want) {
				t.Fatalf("got %v; want %v", h, want)
			}

			bdy, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			got := string(bdy)
			if got != c.want.body {
				t.Fatalf("got %q; want %q", got, c.want.body)
			}
		})
	}
}

func TestAutocompleteHandler(t *testing.T) {
	for _, c := range []struct {
		name string
		q    string
		want *response
	}{
		{"basic", "r",
			&response{
				status:   http.StatusOK,
				template: "json",
				data: suggest.Results{
					Suggestions: []string{
						"radiohead",
						"rage against the machine",
						"red hot chili peppers",
						"r.e.m.",
						"rolling stones",
						"rollins band",
						"rusted root",
					},
				},
			},
		},
		{"default !bangs", "!",
			&response{
				status:   http.StatusOK,
				template: "json",
				data: bangs.Results{
					Suggestions: []bangs.Suggestion{
						{Trigger: "g", Name: "Google", FavIcon: "/image/32x,sOidINN3lLWugS3iyX9NlOZNNR9KlRs1eUcqSReZpZ1Y=/https://www.google.com/favicon.ico"},
						{Trigger: "a", Name: "Amazon", FavIcon: "/image/32x,sOqZY1UB3EmwsX6xESH_XgeJzQhOfB30yMNC9aaN1oCA=/https://www.amazon.com/favicon.ico"},
						{Trigger: "b", Name: "Bing", FavIcon: "/image/32x,sLccmz7JmNIZUI-FIfjAwI5ginz62zIQhX_ffuA6DoSQ=/https://www.bing.com/favicon.ico"},
						{Trigger: "reddit", Name: "Reddit", FavIcon: "/image/32x,saRNw35pHUMLQF5-Z8OFhQ1kHz-RioJ7a_S5ZFyTGJF0=/https://www.reddit.com/favicon.ico"},
						{Trigger: "w", Name: "Wikipedia", FavIcon: "/image/32x,szl9NPdfHe0jt93aiLlox2zOB1DX2ThfpEHiI3AZWUpQ=/https://en.wikipedia.org/favicon.ico"},
					},
				},
			},
		},
		{"g !bangs", "!g",
			&response{
				status:   http.StatusOK,
				template: "json",
				data: bangs.Results{
					Suggestions: []bangs.Suggestion{
						{Trigger: "g", Name: "Google", FavIcon: "/image/32x,sOidINN3lLWugS3iyX9NlOZNNR9KlRs1eUcqSReZpZ1Y=/https://www.google.com/favicon.ico"},
						{Trigger: "gh", Name: "GitHub", FavIcon: "/image/32x,stnNTL-BiRf_CwEuaKJfpgC1xRR8is9PqSW-qLgt3J-s=/https://github.com/favicon.ico"},
					},
				},
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Suggest: &mockSuggester{},
			}

			var err error

			f.Bangs, err = bangsFromConfig()
			if err != nil {
				t.Fatal(err)
			}
			f.Bangs.Suggester = &mockBangSuggester{}

			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("q", c.q)
			req.URL.RawQuery = q.Encode()

			got := f.autocompleteHandler(httptest.NewRecorder(), req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestParseTemplates(t *testing.T) {
	ParseTemplates()

	if _, ok := templates["search"]; !ok {
		t.Fatal("Our search template is not in our templates map.")
	}

	if _, ok := templates["about"]; !ok {
		t.Fatal("Our about template is not in our templates map.")
	}
}

type mockSuggester struct {
	ex bool
}

func (ms *mockSuggester) Exists(q string) (bool, error) {
	return ms.ex, nil
}

func (ms *mockSuggester) Insert(q string) error {
	return nil
}

func (ms *mockSuggester) Increment(q string) error {
	return nil
}

func (ms *mockSuggester) Completion(q string, size int) (suggest.Results, error) {
	s := suggest.Results{}

	if q == "r" {
		s.Suggestions = []string{
			"radiohead",
			"rage against the machine",
			"red hot chili peppers",
			"r.e.m.",
			"rolling stones",
			"rollins band",
			"rusted root",
		}
	}
	return s, nil
}

func (ms *mockSuggester) IndexExists() (bool, error) {
	return ms.ex, nil
}

func (ms *mockSuggester) Setup() error { return nil }

type mockBangSuggester struct{}

func (m *mockBangSuggester) SuggestResults(term string, size int) (bangs.Results, error) {
	res := bangs.Results{
		Suggestions: []bangs.Suggestion{
			{Trigger: "g", Name: "Google"},
			{Trigger: "gh", Name: "GitHub"},
		},
	}

	return res, nil
}

func (m *mockBangSuggester) IndexExists() (bool, error) {
	return false, nil
}

func (m *mockBangSuggester) DeleteIndex() error {
	return nil
}

func (m *mockBangSuggester) Setup(bangs []bangs.Bang) error {
	return nil
}
