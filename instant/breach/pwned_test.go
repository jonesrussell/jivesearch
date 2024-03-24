package breach

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		account string
	}

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want *Response
	}{
		{
			name: "basic",
			u:    `https://haveibeenpwned.com/api/v2/breachedaccount/test@example.com`,
			resp: `[{"Title":"000webhost","Name":"000webhost","Domain":"000webhost.com","BreachDate":"2015-03-01","AddedDate":"2015-10-26T23:35:45Z","ModifiedDate":"2017-12-10T21:44:27Z","PwnCount":14936670,"Description":"Some description here.","DataClasses":["Email addresses","IP addresses","Names","Passwords"],"IsVerified":true,"IsFabricated":false,"IsSensitive":false,"IsActive":true,"IsRetired":false,"IsSpamList":false,"LogoType":"png"},{"Title":"8tracks","Name":"8tracks","Domain":"8tracks.com","BreachDate":"2017-06-27","AddedDate":"2018-02-16T07:09:30Z","ModifiedDate":"2018-02-16T07:09:30Z","PwnCount":7990619,"Description":"Another description here.","DataClasses":["Email addresses","Passwords"],"IsVerified":true,"IsFabricated":false,"IsSensitive":false,"IsActive":true,"IsRetired":false,"IsSpamList":false,"LogoType":"png"}]`,
			args: args{
				account: "test@example.com",
			},
			want: &Response{
				Account: "test@example.com",
				Breaches: []Breach{
					{
						Name:        "000webhost",
						Domain:      "000webhost.com",
						Date:        time.Date(2015, 3, 1, 0, 0, 0, 0, time.UTC),
						Count:       14936670,
						Description: "Some description here.",
						Items:       []string{"Email addresses", "IP addresses", "Names", "Passwords"},
					},
					{
						Name:        "8tracks",
						Domain:      "8tracks.com",
						Date:        time.Date(2017, 6, 27, 0, 0, 0, 0, time.UTC),
						Count:       7990619,
						Description: "Another description here.",
						Items:       []string{"Email addresses", "Passwords"},
					},
				},
				Provider: HaveIBeenPwnedProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder) // no responder found????

			p := &Pwned{
				HTTPClient: &http.Client{},
			}

			got, err := p.Fetch(tt.args.account)
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
