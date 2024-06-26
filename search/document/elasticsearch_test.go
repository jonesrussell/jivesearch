package document

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/olivere/elastic/v7"
	"golang.org/x/text/language"
)

func TestAnalyzer(t *testing.T) {
	for _, c := range []struct {
		name string
		lang language.Tag
		want string
	}{
		{"English", language.English, "english"},
		{"British English", language.BritishEnglish, "english"},
		{"Spanish", language.Spanish, "spanish"},
		{"European Spanish", language.EuropeanSpanish, "spanish"},
		{"Latin American Spanish", language.LatinAmericanSpanish, "spanish"},
		{"German", language.German, "german"},
		{"Portuguese", language.Portuguese, "portuguese"},
		{"European Portuguese", language.EuropeanPortuguese, "portuguese"},
		{"Brazilian Portuguese", language.BrazilianPortuguese, "brazilian"},
	} {
		t.Run(c.name, func(t *testing.T) {
			handler := http.NotFound
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			}))
			defer ts.Close()

			handler = func(w http.ResponseWriter, r *http.Request) {}

			e, err := MockService(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			got, err := e.Analyzer(c.lang)
			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("got %q; want %q", got, c.want)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		resp   string
	}{
		{
			name:   "ok",
			status: http.StatusOK,
			resp:   `{"acknowledged": true}`,
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			handler := http.NotFound
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handler(w, r)
			}))
			defer ts.Close()

			handler = func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(c.status)
				if _, err := w.Write([]byte(c.resp)); err != nil {
					t.Fatal(err)
				}
			}

			e, err := MockService(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			if err := e.Setup(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func MockService(url string) (*ElasticSearch, error) {
	client, err := elastic.NewSimpleClient(elastic.SetURL(url))
	if err != nil {
		return nil, err
	}

	return &ElasticSearch{
		Client: client, Index: "search", Type: "document",
	}, nil
}
