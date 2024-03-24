package congress

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestFetchMembers(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		key   string
		state string
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
			u:    `https://api.propublica.org/congress/v1/members/house/UT/current.json`,
			resp: `{"status":"OK","copyright":"Copyright (c) 2018 Pro Publica Inc. All Rights Reserved.","results":[{"id":"B001250","name":"Rob Bishop","first_name":"Rob","middle_name":null,"last_name":"Bishop","suffix":null,"role":"Representative","gender":"M","party":"R","times_topics_url":null,"twitter_id":"RepRobBishop","facebook_account":"RepRobBishop","youtube_id":"CongressmanBishop","seniority":"16","next_election":"2018","api_uri":"https://api.propublica.org/congress/v1/members/B001250.json","district":"1","at_large":false},{"id":"S001192","name":"Chris Stewart","first_name":"Chris","middle_name":null,"last_name":"Stewart","suffix":null,"role":"Representative","gender":"M","party":"R","times_topics_url":null,"twitter_id":"RepChrisStewart","facebook_account":"RepChrisStewart","youtube_id":"repchrisstewart","seniority":"6","next_election":"2018","api_uri":"https://api.propublica.org/congress/v1/members/S001192.json","district":"2","at_large":false},{"id":"C001114","name":"John Curtis","first_name":"John","middle_name":null,"last_name":"Curtis","suffix":null,"role":"Representative","gender":"M","party":"R","times_topics_url":"","twitter_id":"RepJohnCurtis","facebook_account":null,"youtube_id":null,"seniority":"2","next_election":"2018","api_uri":"https://api.propublica.org/congress/v1/members/C001114.json","district":"3","at_large":false},{"id":"L000584","name":"Mia Love","first_name":"Mia","middle_name":null,"last_name":"Love","suffix":null,"role":"Representative","gender":"F","party":"R","times_topics_url":null,"twitter_id":"repmialove","facebook_account":null,"youtube_id":null,"seniority":"4","next_election":"2018","api_uri":"https://api.propublica.org/congress/v1/members/L000584.json","district":"4","at_large":false}]}`,
			args: args{
				key:   "some_key",
				state: "utah",
			},
			want: &Response{
				Location: &Location{
					Short: "UT",
					State: "Utah",
				},
				Role: House,
				Members: []Member{
					{
						Name:         "Rob Bishop",
						District:     1,
						Gender:       "M",
						Party:        "R",
						Twitter:      "RepRobBishop",
						Facebook:     "RepRobBishop",
						NextElection: 2018,
					},
					{
						Name:         "Chris Stewart",
						District:     2,
						Gender:       "M",
						Party:        "R",
						Twitter:      "RepChrisStewart",
						Facebook:     "RepChrisStewart",
						NextElection: 2018,
					},
					{
						Name:         "John Curtis",
						District:     3,
						Gender:       "M",
						Party:        "R",
						Twitter:      "RepJohnCurtis",
						Facebook:     "",
						NextElection: 2018,
					},
					{
						Name:         "Mia Love",
						District:     4,
						Gender:       "F",
						Party:        "R",
						Twitter:      "repmialove",
						Facebook:     "",
						NextElection: 2018,
					},
				},
				Provider: ProPublicaProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder) // no responder found????

			p := &ProPublica{
				Key:        tt.args.key,
				HTTPClient: &http.Client{},
			}

			loc := ValidateState(tt.args.state)

			got, err := p.FetchMembers(loc)
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

func TestFetchSenators(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		key   string
		state string
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
			u:    `https://api.propublica.org/congress/v1/members/senate/UT/current.json`,
			resp: `{"status":"OK","copyright":"Copyright (c) 2018 Pro Publica Inc. All Rights Reserved.","results":[{"id":"H000338","name":"Orrin G. Hatch","first_name":"Orrin","middle_name":"G.","last_name":"Hatch","suffix":null,"role":"Senator, 1st Class","gender":"M","party":"R","times_topics_url":"http:\/\/topics.nytimes.com\/top\/reference\/timestopics\/people\/h\/orrin_g_hatch\/index.html","twitter_id":"SenOrrinHatch","facebook_account":"senatororrinhatch","youtube_id":"SenatorOrrinHatch","seniority":"41","next_election":"2018","api_uri":"https:\/\/api.propublica.org\/congress\/v1\/members\/H000338.json"},{"id":"L000577","name":"Mike Lee","first_name":"Mike","middle_name":null,"last_name":"Lee","suffix":null,"role":"Senator, 3rd Class","gender":"M","party":"R","times_topics_url":"","twitter_id":"SenMikeLee","facebook_account":"senatormikelee","youtube_id":"senatormikelee","seniority":"7","next_election":"2022","api_uri":"https:\/\/api.propublica.org\/congress\/v1\/members\/L000577.json"}]}`,
			args: args{
				key:   "some_key",
				state: "utah",
			},
			want: &Response{
				Location: &Location{
					Short: "UT",
					State: "Utah",
				},
				Role: Senators,
				Members: []Member{
					{
						Name:         "Orrin G. Hatch",
						Gender:       "M",
						Party:        "R",
						Twitter:      "SenOrrinHatch",
						Facebook:     "senatororrinhatch",
						NextElection: 2018,
					},
					{
						Name:         "Mike Lee",
						Gender:       "M",
						Party:        "R",
						Twitter:      "SenMikeLee",
						Facebook:     "senatormikelee",
						NextElection: 2022,
					},
				},
				Provider: ProPublicaProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder) // no responder found????

			p := &ProPublica{
				Key:        tt.args.key,
				HTTPClient: &http.Client{},
			}

			loc := ValidateState(tt.args.state)

			got, err := p.FetchSenators(loc)
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
