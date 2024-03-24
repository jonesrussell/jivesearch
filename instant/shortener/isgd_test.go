package shortener

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestIsGdService(t *testing.T) {
	var parseURL = func(u string) *url.URL {
		p, _ := url.Parse(u)
		return p
	}

	type args struct {
		original string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		url  string
		raw  string
		want *Response
	}{
		{
			name: "shorten www.example.com",
			args: args{"example.com"},
			url:  "https://is.gd/create.php?format=json&url=example.com",
			raw:  `{"shorturl": "https://is.gd/fsdjfklsd"}`,
			want: &Response{
				Original: parseURL("example.com"),
				Short:    parseURL("https://is.gd/fsdjfklsd"),
				Provider: IsGdProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", tt.url, responder)

			g := &IsGd{
				HTTPClient: &http.Client{},
			}
			got, err := g.Shorten(parseURL(tt.args.original))
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
