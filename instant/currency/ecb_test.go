package currency

import (
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestECBFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
		want *Response
	}{
		{
			name: "basic",
			u:    `http://www.ecb.europa.eu/stats/eurofxref/eurofxref-hist-90d.xml`,
			resp: `<?xml version="1.0" encoding="UTF-8"?><gesmes:Envelope xmlns:gesmes="http://www.gesmes.org/xml/2002-08-01" xmlns="http://www.ecb.int/vocabulary/2002-08-01/eurofxref"><gesmes:subject>Reference rates</gesmes:subject><gesmes:Sender><gesmes:name>European Central Bank</gesmes:name></gesmes:Sender><Cube><Cube time="2018-08-08"><Cube currency="USD" rate="1.1589"/><Cube currency="JPY" rate="128.72"/><Cube currency="GBP" rate="0.90085"/></Cube><Cube time="2018-08-07"><Cube currency="USD" rate="1.1602"/><Cube currency="JPY" rate="128.88"/><Cube currency="GBP" rate="0.89483"/></Cube><Cube time="2018-08-03"><Cube currency="USD" rate="1.1588"/><Cube currency="JPY" rate="129.3"/><Cube currency="GBP" rate="0.8905"/></Cube></Cube></gesmes:Envelope>`,
			want: &Response{
				Base: USD,
				History: map[string][]*Rate{
					EUR.Short: {
						{
							DateTime: time.Date(2018, 8, 8, 0, 0, 0, 0, time.UTC),
							Rate:     1.1589,
						},
						{
							DateTime: time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
							Rate:     1.1602,
						},
						{
							DateTime: time.Date(2018, 8, 3, 0, 0, 0, 0, time.UTC),
							Rate:     1.1588,
						},
					},
					USD.Short: {
						{
							DateTime: time.Date(2018, 8, 8, 0, 0, 0, 0, time.UTC),
							Rate:     1,
						},
						{
							DateTime: time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
							Rate:     1,
						},
						{
							DateTime: time.Date(2018, 8, 3, 0, 0, 0, 0, time.UTC),
							Rate:     1,
						},
					},
					JPY.Short: {
						{
							DateTime: time.Date(2018, 8, 8, 0, 0, 0, 0, time.UTC),
							Rate:     .009,
						},
						{
							DateTime: time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
							Rate:     .009,
						},
						{
							DateTime: time.Date(2018, 8, 3, 0, 0, 0, 0, time.UTC),
							Rate:     .009,
						},
					},
					GBP.Short: {
						{
							DateTime: time.Date(2018, 8, 8, 0, 0, 0, 0, time.UTC),
							Rate:     1.2865,
						},
						{
							DateTime: time.Date(2018, 8, 7, 0, 0, 0, 0, time.UTC),
							Rate:     1.2966,
						},
						{
							DateTime: time.Date(2018, 8, 3, 0, 0, 0, 0, time.UTC),
							Rate:     1.3013,
						},
					},
				},
				ForexProvider: ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			ecb := &ECB{}
			got, err := ecb.Fetch()
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
