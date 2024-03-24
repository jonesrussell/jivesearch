package parcel

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestUSPSFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
	}{
		{
			name: "9374889701090078857768",
			u:    `http://production.shippingapis.com/ShippingAPI.dll`,
			resp: `<?xml version="1.0"?><TrackResponse><TrackInfo ID="9374889701090078857768"><TrackSummary><EventTime>1:57 pm</EventTime><EventDate>March 12, 2018</EventDate><Event>Delivered</Event><EventCity>Some City</EventCity><EventState>ID</EventState><EventZIPCode>90210</EventZIPCode><EventCountry/><FirmName/><Name/><AuthorizedAgent>false</AuthorizedAgent><DeliveryAttributeCode>01</DeliveryAttributeCode></TrackSummary><TrackDetail><EventTime>8:13 am</EventTime><EventDate>March 14, 2018</EventDate><Event>Out for Delivery</Event><EventCity>Close to Some City</EventCity><EventState>ID</EventState><EventZIPCode>90211</EventZIPCode><EventCountry/><FirmName/><Name/><AuthorizedAgent>false</AuthorizedAgent></TrackDetail><TrackDetail><EventTime>7:11 am</EventTime><EventDate>March 14, 2018</EventDate><Event>Almost there dude</Event><EventCity>Almost</EventCity><EventState>ID</EventState><EventZIPCode>90209</EventZIPCode><EventCountry/><FirmName/><Name/><AuthorizedAgent>false</AuthorizedAgent></TrackDetail></TrackInfo></TrackResponse>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			a := &USPS{
				HTTPClient: &http.Client{},
			}

			got, err := a.Fetch(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			want := Response{
				TrackingNumber: strings.ToUpper(tt.name),
				Updates: []Update{
					{
						DateTime: time.Date(2018, 3, 12, 13, 57, 0, 0, time.UTC),
						Location: Location{
							City: "Some City", State: "ID", Country: "",
						},
						Status: "Delivered",
					},
					{
						DateTime: time.Date(2018, 3, 14, 8, 13, 0, 0, time.UTC),
						Location: Location{
							City: "Close to Some City", State: "ID", Country: "",
						},
						Status: "Out for Delivery",
					},
					{
						DateTime: time.Date(2018, 3, 14, 7, 11, 0, 0, time.UTC),
						Location: Location{
							City: "Almost", State: "ID", Country: "",
						},
						Status: "Almost there dude",
					},
				},
				URL: fmt.Sprintf("https://tools.usps.com/go/TrackConfirmAction?origTrackNum=%v", strings.ToUpper(tt.name)),
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}
