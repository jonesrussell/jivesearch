package crawler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jonesrussell/jivesearch/search/document"

	"github.com/olivere/elastic/v7"
)

func TestUpsert(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		resp   string
		doc    *document.Document
		err    error
	}{
		{
			name:   "basic",
			status: http.StatusCreated,
			resp: `{
			  "took": 27,
			  "errors": false,
			  "items": [
					{
			      "create": {
			        "_index": "search",
			        "_type": "document",
			        "_id": "AVhRlxyshqP4iSOLLnUz",
			        "_version": 1,
			        "_shards": {
			          "total": 2,
			          "successful": 1,
			          "failed": 0
			        },
			        "status": 201
			      }
				  }
				]
			}`,
			doc: &document.Document{
				ID:     "http://www.example.com/path/to/nowhere",
				Domain: "example.com",
				Host:   "http://www.example.com",
			},
			err: nil,
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

			if err := e.Upsert(c.doc); err != nil {
				t.Fatal(err)
			}

			if err := e.Bulk.Flush(); err != nil {
				t.Fatal(err)
			}

			stats := e.Bulk.Stats()
			if stats.Succeeded != 1 {
				t.Fatalf("upsert failed: got %d", stats.Succeeded)
			}
		})
	}
}

func MockService(url string) (*ElasticSearch, error) {
	client, err := elastic.NewSimpleClient(elastic.SetURL(url))
	if err != nil {
		return nil, err
	}

	bulk, err := client.BulkProcessor().Stats(true).Do(context.TODO())
	if err != nil {
		return nil, err
	}

	return &ElasticSearch{
		ElasticSearch: &document.ElasticSearch{
			Client: client, Index: "search", Type: "document",
		},
		Bulk: bulk,
	}, nil
}
