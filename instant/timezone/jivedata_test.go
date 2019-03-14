package timezone

import (
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestJiveDataFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		lat float64
		lon float64
	}

	for _, tt := range []struct {
		name string
		args
		want string
		raw  string
	}{
		{
			name: "Sydney, Australia",
			args: args{-33.8667, 151.2},
			want: "Australia/Sydney",
			raw:  `{"timezone":"Australia/Sydney"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			u, err := url.Parse("https://jivedata.com/timezone")
			if err != nil {
				t.Fatal(err)
			}

			q := u.Query()
			q.Set("key", j.Key)
			q.Set("lat", strconv.FormatFloat(tt.args.lat, 'f', -1, 64))
			q.Set("lon", strconv.FormatFloat(tt.args.lon, 'f', -1, 64))
			u.RawQuery = q.Encode()

			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", u.String(), responder)

			got, err := j.Fetch(tt.args.lat, tt.args.lon)
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
