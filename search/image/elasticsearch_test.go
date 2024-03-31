package image

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/olivere/elastic/v7"
)

func TestFetch(t *testing.T) {
	type want struct {
		*Results
		err error
	}

	for _, c := range []struct {
		name   string
		query  string
		safe   bool
		number int
		offset int
		status int
		resp   string
		want
	}{
		{
			name:   "basic",
			query:  "Bob Dylan",
			safe:   true,
			number: 25,
			offset: 1,
			status: http.StatusOK,
			resp: `{
				"took": 2,
				"timed_out": false,
				"_shards": {
				  "total": 5,
				  "successful": 5,
				  "skipped": 0,
				  "failed": 0
				},
				"hits": {
				  "total": 16077,
				  "max_score": 1,
				  "hits": [
					{
					  "_index": "test-images",
					  "_type": "image",
					  "_id": "https://lh3.googleusercontent.com/QW9dvq_v2f4ECFfEG-vN-5Ex8wZkllBb-ORVlbSwnmXjGKuYiB-3C5kfnpDXEMXBr-S0XXTumABojH6BoA=pf-w200-h200",
					  "_score": 1,
					  "_source": {
						"id": "https://lh3.googleusercontent.com/QW9dvq_v2f4ECFfEG-vN-5Ex8wZkllBb-ORVlbSwnmXjGKuYiB-3C5kfnpDXEMXBr-S0XXTumABojH6BoA=pf-w200-h200",
						"domain": "googleusercontent.com",
						"alt": "Russian journalist and Kremlin critic Arkady Babchenko killed in Kiev | The Independent",
						"nsfw_score": 0.1410432904958725,
						"crawled": "20180529",
						"width": 200,
						"height": 200
					  }
					},
					{
					  "_index": "test-images",
					  "_type": "image",
					  "_id": "https://static.xx.fbcdn.net/rsrc.php/v3/yJ/r/yLtEhZl0QOJ.png",
					  "_score": 1,
					  "_source": {
						"id": "https://static.xx.fbcdn.net/rsrc.php/v3/yJ/r/yLtEhZl0QOJ.png",
						"domain": "fbcdn.net",
						"alt": "Highlights info row image",
						"nsfw_score": 0.007919724099338055,
						"crawled": "20180529",
						"width": 24,
						"height": 24
					  }
					}
				  ]
				}
			}`,
			want: want{
				&Results{
					Count: 16077,
					Images: []*Image{
						{
							ID:      "https://lh3.googleusercontent.com/QW9dvq_v2f4ECFfEG-vN-5Ex8wZkllBb-ORVlbSwnmXjGKuYiB-3C5kfnpDXEMXBr-S0XXTumABojH6BoA=pf-w200-h200",
							Domain:  "googleusercontent.com",
							Alt:     "Russian journalist and Kremlin critic Arkady Babchenko killed in Kiev | The Independent",
							NSFW:    0.1410432904958725,
							Width:   200,
							Height:  200,
							EXIF:    EXIF{},
							Crawled: "20180529",
						},
						{
							ID:      "https://static.xx.fbcdn.net/rsrc.php/v3/yJ/r/yLtEhZl0QOJ.png",
							Domain:  "fbcdn.net",
							Alt:     "Highlights info row image",
							NSFW:    0.007919724099338055,
							Width:   24,
							Height:  24,
							EXIF:    EXIF{},
							Crawled: "20180529",
						},
					},
				},
				nil,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			handler := http.NotFound
			ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(c.status)
				handler(w, r)
			}))
			defer ts.Close()

			handler = func(w http.ResponseWriter, _ *http.Request) {
				if _, err := w.Write([]byte(c.resp)); err != nil {
					t.Fatal(err)
				}
			}

			e, err := MockService(ts.URL)
			if err != nil {
				t.Fatal(err)
			}

			got, err := e.Fetch(c.query, c.safe, c.number, c.offset)
			if err != c.want.err {
				t.Fatalf("got err %q; want %q", err, c.want.err)
			}

			if !reflect.DeepEqual(got, c.want.Results) {
				t.Fatalf("got %+v; want %+v", got, c.want.Results)
			}
		})
	}
}

func TestUpsert(t *testing.T) {
	for _, c := range []struct {
		name   string
		status int
		resp   string
		img    *Image
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
			        "_index": "images",
			        "_type": "images",
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
			img: &Image{
				ID:      "http://www.example.com/path/to/nowhere",
				Alt:     "some image we have here",
				NSFW:    .0002,
				Crawled: "20180520",
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

			if err := e.Upsert(c.img); err != nil {
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

func TestUncrawled(t *testing.T) {
	type args struct {
		number int
		since  time.Time
	}

	type want struct {
		images []*Image
		err    error
	}

	for _, c := range []struct {
		name string
		args
		status int
		resp   string
		want
	}{
		{
			name: "basic",
			args: args{
				100, time.Date(2018, 5, 2, 0, 0, 0, 0, time.UTC),
			},
			status: http.StatusCreated,
			resp: `{
				"took": 24,
				"timed_out": false,
				"_shards": {
				  "total": 5,
				  "successful": 5,
				  "skipped": 0,
				  "failed": 0
				},
				"hits": {
				  "total": 58723,
				  "max_score": 0,
				  "hits": []
				},
				"aggregations": {
				  "by_domain": {
					"doc_count_error_upper_bound": 0,
					"sum_other_doc_count": 0,
					"buckets": [
					  {
						"key": "wikimedia.org",
						"doc_count": 52365,
						"get_one": {
						  "hits": {
							"total": 52365,
							"max_score": 1,
							"hits": [
							  {
								"_index": "test-images",
								"_type": "image",
								"_id": "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b3/HumanWhistle.jpg/120px-HumanWhistle.jpg",
								"_score": 1,
								"_source": {
								  "domain": "wikimedia.org",
								  "alt": "HumanWhistle.jpg",
								  "id": "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b3/HumanWhistle.jpg/120px-HumanWhistle.jpg"
								}
							  }
							]
						  }
						}
					  },
					  {
						"key": "nih.gov",
						"doc_count": 1353,
						"get_one": {
						  "hits": {
							"total": 1353,
							"max_score": 1.4e-45,
							"hits": [
							  {
								"_index": "test-images",
								"_type": "image",
								"_id": "https://www.nih.gov/sites/default/files/styles/slide_breakpoint-medium-small/public/home_0/slides/current/slide-gluten-free.jpg?itok=KSL9IXCf&timestamp=1526918919",
								"_score": 0,
								"_source": {
								  "domain": "nih.gov",
								  "alt": "Loaves of bread in a crate with &quot;Gluten Free&quot; printed on the side of it",
								  "id": "https://www.nih.gov/sites/default/files/styles/slide_breakpoint-medium-small/public/home_0/slides/current/slide-gluten-free.jpg?itok=KSL9IXCf&timestamp=1526918919"
								}
							  }
							]
						  }
						}
					  }
					]
					}
				}
			}`,
			want: want{
				images: []*Image{
					{
						ID:     "https://upload.wikimedia.org/wikipedia/commons/thumb/b/b3/HumanWhistle.jpg/120px-HumanWhistle.jpg",
						Domain: "wikimedia.org",
						Alt:    "HumanWhistle.jpg",
					},
					{
						ID:     "https://www.nih.gov/sites/default/files/styles/slide_breakpoint-medium-small/public/home_0/slides/current/slide-gluten-free.jpg?itok=KSL9IXCf&timestamp=1526918919",
						Domain: "nih.gov",
						Alt:    "Loaves of bread in a crate with &quot;Gluten Free&quot; printed on the side of it",
					},
				},
				err: nil,
			},
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

			got, err := e.Uncrawled(c.args.number, c.args.since)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want.images) {
				t.Fatalf("got %+v; want %+v", got, c.want.images)
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

	bulk, err := client.BulkProcessor().Stats(true).Do(context.TODO())
	if err != nil {
		return nil, err
	}

	return &ElasticSearch{
		Client: client,
		Index:  "images",
		Type:   "image",
		Bulk:   bulk,
	}, nil
}
