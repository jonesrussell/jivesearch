package musicbrainz

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/jivesearch/jivesearch/instant/discography"
)

func TestJiveDataFetch(t *testing.T) {
	type args struct {
		artist string
		u      []string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		raw  string
		want []discography.Album
	}{
		{
			name: "metallica discography",
			args: args{
				artist: "metallica discography",
				u: []string{
					"http://coverartarchive.org/release/a89e1d92-5381-4dab-ba51-733137d0e431/15674154080-250.jpg",
					"http://coverartarchive.org/release/589ff96d-0be8-3f82-bdd2-299592e51b40/15674886619-250.jpg",
				},
			},
			raw: `[
				{
					"Name": "Kill \u2019em All",
					"Published": "1983-07-25T00:00:00Z",
					"ID": "",
					"URL": {
						"Scheme": "http",
						"Opaque": "",
						"User": null,
						"Host": "coverartarchive.org",
						"Path": "\/release\/a89e1d92-5381-4dab-ba51-733137d0e431\/15674154080-250.jpg",
						"RawPath": "",
						"ForceQuery": false,
						"RawQuery": "",
						"Fragment": ""
					}
				},
				{
					"Name": "Ride the Lightning",
					"Published": "1984-07-30T00:00:00Z",
					"ID": "",
					"URL": {
						"Scheme": "http",
						"Opaque": "",
						"User": null,
						"Host": "coverartarchive.org",
						"Path": "\/release\/589ff96d-0be8-3f82-bdd2-299592e51b40\/15674886619-250.jpg",
						"RawPath": "",
						"ForceQuery": false,
						"RawQuery": "",
						"Fragment": ""
					}
				}
			]`,
			want: []discography.Album{
				{
					Name:      "Kill \u2019em All",
					Published: time.Date(1983, 7, 25, 0, 0, 0, 0, time.UTC),
					Image:     discography.Image{},
				},
				{
					Name:      "Ride the Lightning",
					Published: time.Date(1984, 7, 30, 0, 0, 0, 0, time.UTC),
					Image:     discography.Image{},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			u, err := url.Parse("https://jivedata.com/musicbrainz")
			if err != nil {
				t.Fatal(err)
			}

			q := u.Query()
			q.Set("key", j.Key)
			q.Set("artist", tt.args.artist)
			u.RawQuery = q.Encode()

			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", u.String(), responder)

			got, err := j.Fetch(tt.args.artist)
			if err != nil {
				t.Fatal(err)
			}

			for i, uu := range tt.args.u {
				uu, err := url.Parse(uu)
				if err != nil {
					t.Fatal(err)
				}
				tt.want[i].Image.URL = uu
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
