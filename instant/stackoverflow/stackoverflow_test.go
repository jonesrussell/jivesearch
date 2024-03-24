package stackoverflow

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestAPIFetch(t *testing.T) {
	type args struct {
		query string
		tags  []string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want Response
	}{
		{
			name: "php loop",
			args: args{"loop", []string{"php"}},
			u:    `https://api.stackexchange.com/2.2/search/advanced?answers=1&filter=%21OfZYd4zGqhN8IapZI6RQ6uaya_ZBeR7bHr1c%29NV5Cu9&key=&order=desc&page=1&pagesize=1&q=loop&site=stackoverflow&sort=relevance&tagged=php`,
			resp: `{
				"items": [
					{
						"answers": [
							{
								"owner": {
								"display_name": "NikiC"
								},
								"score": 1273,
								"body": "an answer"					  
							}
						],
						"link": "https:\/\/stackoverflow.com\/questions\/10057671\/how-does-php-foreach-actually-work",
						"title": "How does PHP &#39;foreach&#39; actually work?"
					}
				],
				"quota_max": 300,
				"quota_remaining": 197
			}`,
			want: Response{
				Items: []Item{
					{
						Answers: []Answer{
							{
								Owner: Owner{
									DisplayName: "NikiC",
								},
								Score: 1273,
								Body:  "an answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/10057671/how-does-php-foreach-actually-work",
						Title: "How does PHP &#39;foreach&#39; actually work?",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			a := &API{
				Key:        "",
				HTTPClient: &http.Client{},
			}
			got, err := a.Fetch(tt.args.query, tt.args.tags)
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
