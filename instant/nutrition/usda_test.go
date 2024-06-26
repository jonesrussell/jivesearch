package nutrition

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

type FoodDetail struct {
	SR        string        `json:"sr"`
	Type      string        `json:"type"`
	Desc      Description   `json:"desc"`
	Nutrients []Nutrient    `json:"nutrients"`
	Sources   []interface{} `json:"sources"`
	Footnotes []interface{} `json:"footnotes"`
	Langual   []interface{} `json:"langual"`
}

type Description struct {
	Ndbno string  `json:"ndbno"`
	Name  string  `json:"name"`
	SD    string  `json:"sd"`
	FG    string  `json:"fg"`
	SN    string  `json:"sn"`
	CN    string  `json:"cn"`
	Manu  string  `json:"manu"`
	NF    float64 `json:"nf"`
	CF    float64 `json:"cf"`
	FF    float64 `json:"ff"`
	PF    float64 `json:"pf"`
	R     string  `json:"r"`
	RD    string  `json:"rd"`
	DS    string  `json:"ds"`
	RU    string  `json:"ru"`
}

func TestUSDALookup(t *testing.T) {
	type args struct {
		q string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want []*ItemResponse
	}{
		{
			name: "basic",
			args: args{"egg"},
			u:    `https://api.nal.usda.gov/ndb/search/?api_key=&format=json&max=25&offset=0&q=egg&sort=n`,
			resp: `{
				"list": {
					"q": "egg",
					"sr": "1",
					"ds": "any",
					"start": 0,
					"end": 25,
					"total": 1956,
					"group": "",
					"sort": "n",
					"item": [
						{
							"offset": 0,
							"group": "Dairy",
							"name": "egg",
							"ndbno": "12",
							"ds": "LI",
							"manu": ""
						},
						{
							"offset": 1,
							"group": "Dairy",
							"name": "Some stupid egg",
							"ndbno": "18",
							"ds": "SE",
							"manu": "StupidCo"
						}
					]
				}
			}`,
			want: []*ItemResponse{
				{
					Name:  "egg",
					NDBNO: "12",
				},
				{
					Name:         "Some stupid egg",
					NDBNO:        "18",
					Manufacturer: "StupidCo",
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			a := &USDA{
				Key:        "",
				HTTPClient: &http.Client{},
			}
			got, err := a.Lookup(tt.args.q)
			if err != nil {
				t.Fatal(err)
			}

			for _, g := range got {
				fmt.Printf("%+v\n", g)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
