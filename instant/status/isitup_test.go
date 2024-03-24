package status

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
		want *Response
	}{
		{
			name: "example.com",
			u:    `https://isitup.org/example.com.json`,
			resp: `{
				"domain": "example.com",
				"port": 80,
				"status_code": 1,
				"response_ip": "1.2.3.4",
				"response_code": 200,
				"response_time": 0.002
			}`,
			want: &Response{
				Domain:   "example.com",
				Port:     80,
				Status:   1,
				IP:       "1.2.3.4",
				Code:     200,
				Time:     0.002,
				Provider: IsItUpProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder) // no responder found????

			p := &IsItUp{
				HTTPClient: &http.Client{},
			}

			got, err := p.Fetch(tt.name)
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
