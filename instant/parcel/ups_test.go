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

func TestUPSFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
		want Response
	}{
		{
			name: "1Z12345E6605272234",
			u:    `https://onlinetools.ups.com/rest/Track`,
			resp: `{"TrackResponse":{"Response":{"ResponseStatus":{"Code":"1", "Description":"Success"}, "TransactionReference":{"CustomerContext":"Your Test Case Summary Description"}}, "Shipment":{"InquiryNumber":{"Code":"01", "Description":"ShipmentIdentificationNumber", "Value":"1Z12345E6605272234"}, "ShipperNumber":"", "ShipmentAddress":[{"Type":{"Code":"01", "Description":"Shipper Address"}, "Address":{"AddressLine":["1 Main", "Unit 10"], "City":"Some City", "StateProvinceCode":"ID", "PostalCode":"90210", "CountryCode":"US"}}, {"Type":{"Code":"02", "Description":"ShipTo Address"}, "Address":{"City":"Another City", "StateProvinceCode":"KS", "PostalCode":"90211", "CountryCode":"US"}}], "ShipmentWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"17.50"}, "Service":{"Code":"003", "Description":"UPS GROUND"}, "ReferenceNumber":[{"Code":"01", "Value":""}, {"Code":"01", "Value":""}], "DeliveryDetail":{"Type":{"Code":"03", "Description":"Scheduled Delivery"}, "Date":"20180311"}, "Package":{"TrackingNumber":"1Z12345E6605272234", "PackageServiceOption":{"Type":{"Code":"024", "Description":"ARS"}}, "Activity":[{"ActivityLocation":{"Address":{"City":"Banahana", "StateProvinceCode":"ID", "CountryCode":"US"}}, "Status":{"Type":"I", "Description":"Departure Scan", "Code":"DP"}, "Date":"20180311", "Time":"023800"}], "Message":{"Code":"01", "Description":"On Time"}, "PackageWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"17.50"}, "ReferenceNumber":[{"Code":"01", "Value":""}, {"Code":"01", "Value":""}]}}, "Disclaimer":"You are using UPS tracking service on..."}}`,
			want: Response{
				TrackingNumber: strings.ToUpper("1Z12345E6605272234"),
				Updates: []Update{
					{
						DateTime: time.Date(2018, 3, 11, 2, 38, 0, 0, time.UTC),
						Location: Location{
							City: "Banahana", State: "ID", Country: "US",
						},
						Status: "Departure Scan",
					},
				},
				Expected: Expected{
					Delivery: "Scheduled Delivery",
					Date:     time.Date(2018, 3, 11, 0, 0, 0, 0, time.UTC),
				},
				URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", "1Z12345E6605272234"),
			},
		},
		{
			name: "1Z12345E6705272244",
			u:    `https://onlinetools.ups.com/rest/Track`,
			resp: `{"TrackResponse":{"Response":{"ResponseStatus":{"Code":"1", "Description":"Success"}, "TransactionReference":{"CustomerContext":"Your Test Case Summary Description"}}, "Shipment":{"InquiryNumber":{"Code":"01", "Description":"ShipmentIdentificationNumber", "Value":"1Z12345E6705272244"}, "ShipperNumber":"", "ShipmentAddress":[{"Type":{"Code":"01", "Description":"Shipper Address"}, "Address":{"AddressLine":"21 W Chestnut BLVD", "City":"Another City", "StateProvinceCode":"MD", "PostalCode":"90211", "CountryCode":"US"}}, {"Type":{"Code":"02", "Description":"ShipTo Address"}, "Address":{"City":"Rando City", "StateProvinceCode":"VT", "PostalCode":"90210", "CountryCode":"US"}}], "ShipmentWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"1.30"}, "Service":{"Code":"003", "Description":"UPS GROUND"}, "ReferenceNumber":{"Code":"01", "Value":"1271809"}, "PickupDate":"20180429", "Package":{"TrackingNumber":"1Z12345E6705272244", "Activity":[{"ActivityLocation":{"Address":{"City":"Rando City", "StateProvinceCode":"UT", "PostalCode":"90210", "CountryCode":"US"}, "Code":"ML", "Description":"FRONT DOOR"}, "Status":{"Type":"D", "Description":"Delivered", "Code":"FS"}, "Date":"20180501", "Time":"140300"}, {"ActivityLocation":{"Address":{"City":"Me city", "StateProvinceCode":"ID", "CountryCode":"US"}}, "Status":{"Type":"I", "Description":"Out For Delivery Today", "Code":"OT"}, "Date":"20180501", "Time":"084300"}, {"ActivityLocation":{"Address":{"City":"My city", "StateProvinceCode":"ID", "CountryCode":"US"}}, "Status":{"Type":"I", "Description":"Destination Scan", "Code":"YP"}, "Date":"20180430", "Time":"051600"}, {"ActivityLocation":{"Address":{"CountryCode":"US"}}, "Status":{"Type":"M", "Description":"Order Processed: Ready for UPS", "Code":"MP"}, "Date":"20180429", "Time":"173800"}], "PackageWeight":{"UnitOfMeasurement":{"Code":"LBS"}, "Weight":"1.30"}, "ReferenceNumber":[{"Code":"01", "Value":"some value"}]}}}}`,
			want: Response{
				TrackingNumber: strings.ToUpper("1Z12345E6705272244"),
				Updates: []Update{
					{
						DateTime: time.Date(2018, 5, 1, 14, 3, 0, 0, time.UTC),
						Location: Location{
							City: "Rando City", State: "UT", Country: "US",
						},
						Status: "Delivered",
					},
					{
						DateTime: time.Date(2018, 5, 1, 8, 43, 0, 0, time.UTC),
						Location: Location{
							City: "Me city", State: "ID", Country: "US",
						},
						Status: "Out For Delivery Today",
					},
					{
						DateTime: time.Date(2018, 4, 30, 05, 16, 0, 0, time.UTC),
						Location: Location{
							City: "My city", State: "ID", Country: "US",
						},
						Status: "Destination Scan",
					},
					{
						DateTime: time.Date(2018, 4, 29, 17, 38, 0, 0, time.UTC),
						Location: Location{
							City: "", State: "", Country: "US",
						},
						Status: "Order Processed: Ready for UPS",
					},
				},
				Expected: Expected{},
				URL:      fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", "1Z12345E6705272244"),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("POST", tt.u, responder)

			a := &UPS{
				HTTPClient: &http.Client{},
			}
			got, err := a.Fetch(tt.name)
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
