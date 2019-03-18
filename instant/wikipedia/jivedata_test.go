package wikipedia

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"golang.org/x/text/language"
)

func TestJiveDataFetch(t *testing.T) {
	type args struct {
		q string
		l language.Tag
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		raw  string
		want []*Item
	}{
		{
			name: "shaq",
			args: args{"shaq", language.MustParse("en")},
			raw:  `[{"wikibase_item":"Q169452", "language": "en", "title":"Shaquille O'Neal", "text":"Shaquille O'Neal is a basketball player", "claims":{"image":["https://upload.wikimedia.org/wikipedia/commons/a/a7/Shaqmiami.jpg"]}}]`,
			want: []*Item{
				{
					Wikipedia: Wikipedia{
						ID:       "Q169452",
						Language: "en",
						Title:    "Shaquille O'Neal",
						Text:     "Shaquille O'Neal is a basketball player",
					},
					Wikidata: &Wikidata{
						Claims: &Claims{
							Image: []string{"https://upload.wikimedia.org/wikipedia/commons/a/a7/Shaqmiami.jpg"},
						},
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			u, err := url.Parse("https://jivedata.com/wikipedia")
			if err != nil {
				t.Fatal(err)
			}

			q := u.Query()
			q.Set("key", j.Key)
			q.Set("q", tt.args.q)
			q.Set("l", tt.args.l.String())
			u.RawQuery = q.Encode()

			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", u.String(), responder)

			got, err := j.Fetch(tt.args.q, tt.args.l)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}

func TestJiveDataSetup(t *testing.T) {
	for _, tt := range []struct {
		name string
		want error
	}{
		{
			name: "basic",
			want: nil,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			got := j.Setup()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
