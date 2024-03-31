package bangs

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/olivere/elastic/v7"
)

func TestSuggestES(t *testing.T) {
	type args struct {
		term string
		size int
	}

	for _, c := range []struct {
		name string
		args
		status int
		resp   string
		want   Results
	}{
		{
			"ok", args{"500", 10}, http.StatusOK,
			`{
				"took": 0,
				"timed_out": false,
				"_shards": {
					"total": 5,
					"successful": 5,
					"skipped": 0,
					"failed": 0
				},
				"hits": {
					"total": 0,
					"max_score": 0,
					"hits": []
				},
				"suggest": {
					"bang_suggest": [
						{
							"text": "b",
							"offset": 0,
							"length": 1,
							"options": [
								{
									"text": "g",
									"_index": "test-bangs",
									"_type": "bang",
									"_id": "N9Ing2EBsabfnQRLhudz",
									"_score": 1,
									"_source": {
										"bang_suggest": {
											"input": "g",
											"weight": 0
										}
									}
								},
								{
									"text": "gfr",
									"_index": "test-bangs",
									"_type": "bang",
									"_id": "N9Ing2EBxsbfnQRLhudz",
									"_score": 1,
									"_source": {
										"bang_suggest": {
											"input": "gfr",
											"weight": 0
										}
									}
								},
								{
									"text": "gh",
									"_index": "test-bangs",
									"_type": "bang",
									"_id": "N9In322EBsabfnQRLhudz",
									"_score": 1,
									"_source": {
										"bang_suggest": {
											"input": "gh",
											"weight": 0
										}
									}
								},
								{
									"text": "gi",
									"_index": "test-bangs",
									"_type": "bang",
									"_id": "N9Ing2EBsabfnQyumhudz",
									"_score": 1,
									"_source": {
										"bang_suggest": {
											"input": "gi",
											"weight": 0
										}
									}
								}														
							]
						}
					]
				}
			}`,
			Results{Suggestions: []Suggestion{
				{Trigger: "g"}, {Trigger: "gfr"}, {Trigger: "gh"}, {Trigger: "gi"},
			}},
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

			got, err := e.SuggestResults(c.args.term, c.args.size)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

func TestIndexExists(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		want   bool
	}{
		{
			"exists", http.StatusOK, true,
		},
		{
			"doesn't exist", http.StatusNotFound, false,
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
			}

			e, err := MockService(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			got, err := e.IndexExists()
			if err != nil {
				t.Fatal(err)
			}

			if got != c.want {
				t.Fatalf("got %v; want %v", got, c.want)
			}
		})
	}
}

func TestDeleteIndex(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		resp   string
	}{
		{
			"ok", http.StatusOK, `{"acknowledged": true}`,
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

			if err := e.DeleteIndex(); err != nil {
				t.Fatal(err)
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
			"ok", http.StatusOK, `{"acknowledged": true}`,
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

			b, err := fromConfig()
			if err != nil {
				panic(err)
			}

			if err := e.Setup(b.Bangs); err != nil {
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
	return &ElasticSearch{Client: client, Index: "test-bangs", Type: "bang"}, nil
}
