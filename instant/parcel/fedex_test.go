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

func TestFedExFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    string
		resp string
	}{
		{
			name: "149331877648230",
			u:    `https://ws.fedex.com/web-services`,
			resp: `<?xml version="1.0"?><SOAP-ENV:Envelope xmlns:SOAP-ENV="http://schemas.xmlsoap.org/soap/envelope/"><SOAP-ENV:Header/><SOAP-ENV:Body><TrackReply xmlns="http://fedex.com/ws/track/v14"><HighestSeverity>SUCCESS</HighestSeverity><Notifications><Severity>SUCCESS</Severity><Source>trck</Source><Code>0</Code><Message>Request was successfully processed.</Message><LocalizedMessage>Request was successfully processed.</LocalizedMessage></Notifications><TransactionDetail><CustomerTransactionId>Track By Number_v14</CustomerTransactionId><Localization><LanguageCode>EN</LanguageCode><LocaleCode>US</LocaleCode></Localization></TransactionDetail><Version><ServiceId>trck</ServiceId><Major>14</Major><Intermediate>0</Intermediate><Minor>0</Minor></Version><CompletedTrackDetails><HighestSeverity>SUCCESS</HighestSeverity><Notifications><Severity>SUCCESS</Severity><Source>trck</Source><Code>0</Code><Message>Request was successfully processed.</Message><LocalizedMessage>Request was successfully processed.</LocalizedMessage></Notifications><DuplicateWaybill>false</DuplicateWaybill><MoreData>false</MoreData><TrackDetailsCount>0</TrackDetailsCount><TrackDetails><Notification><Severity>SUCCESS</Severity><Source>trck</Source><Code>0</Code><Message>Request was successfully processed.</Message><LocalizedMessage>Request was successfully processed.</LocalizedMessage></Notification><TrackingNumber>149331877648230</TrackingNumber><TrackingNumberUniqueIdentifier>12017~149331877648230~FDEG</TrackingNumberUniqueIdentifier><StatusDetail><CreationTime>2018-01-03T00:00:00</CreationTime><Code>DL</Code><Description>Delivered</Description><Location><City>Kandy</City><StateOrProvinceCode>ID</StateOrProvinceCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></Location></StatusDetail><CarrierCode>FDXG</CarrierCode><OperatingCompanyOrCarrierDescription>FedEx Ground</OperatingCompanyOrCarrierDescription><OtherIdentifiers><PackageIdentifier><Type>CUSTOMER_REFERENCE</Type><Value>403219</Value></PackageIdentifier></OtherIdentifiers><OtherIdentifiers><PackageIdentifier><Type>PURCHASE_ORDER</Type><Value>SO7837111</Value></PackageIdentifier></OtherIdentifiers><Service><Type>FEDEX_GROUND</Type><Description>FedEx Ground</Description><ShortDescription>FG</ShortDescription></Service><PackageWeight><Units>LB</Units><Value>73.0</Value></PackageWeight><PackageDimensions><Length>42</Length><Width>10</Width><Height>17</Height><Units>IN</Units></PackageDimensions><Packaging>Package</Packaging><PackagingType>YOUR_PACKAGING</PackagingType><PhysicalPackagingType>PACKAGE</PhysicalPackagingType><PackageSequenceNumber>1</PackageSequenceNumber><PackageCount>1</PackageCount><Payments><Classification>TRANSPORTATION</Classification><Type>SHIPPER_ACCOUNT</Type><Description>Shipper</Description></Payments><ShipperAddress><City>Bonham</City><StateOrProvinceCode>MD</StateOrProvinceCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></ShipperAddress><OriginLocationAddress><City>SOME BRANCH</City><StateOrProvinceCode>MD</StateOrProvinceCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></OriginLocationAddress><DatesOrTimes><Type>ACTUAL_DELIVERY</Type><DateOrTimestamp>2018-01-03T11:12:54-07:00</DateOrTimestamp></DatesOrTimes><DatesOrTimes><Type>ACTUAL_PICKUP</Type><DateOrTimestamp>2017-12-27T00:00:00</DateOrTimestamp></DatesOrTimes><DestinationAddress><City>Kandy</City><StateOrProvinceCode>ID</StateOrProvinceCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></DestinationAddress><ActualDeliveryAddress><City>Kandy</City><StateOrProvinceCode>ID</StateOrProvinceCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></ActualDeliveryAddress><DeliveryAttempts>0</DeliveryAttempts><DeliverySignatureName>OOFFICE</DeliverySignatureName><TotalUniqueAddressCountInConsolidation>0</TotalUniqueAddressCountInConsolidation><AvailableImages><Type>SIGNATURE_PROOF_OF_DELIVERY</Type></AvailableImages><NotificationEventsAvailable>ON_DELIVERY</NotificationEventsAvailable><DeliveryOptionEligibilityDetails><Option>INDIRECT_SIGNATURE_RELEASE</Option><Eligibility>INELIGIBLE</Eligibility></DeliveryOptionEligibilityDetails><DeliveryOptionEligibilityDetails><Option>REDIRECT_TO_HOLD_AT_LOCATION</Option><Eligibility>INELIGIBLE</Eligibility></DeliveryOptionEligibilityDetails><DeliveryOptionEligibilityDetails><Option>REROUTE</Option><Eligibility>INELIGIBLE</Eligibility></DeliveryOptionEligibilityDetails><DeliveryOptionEligibilityDetails><Option>RESCHEDULE</Option><Eligibility>INELIGIBLE</Eligibility></DeliveryOptionEligibilityDetails><Events><Timestamp>2018-01-03T11:12:45-07:00</Timestamp><EventType>DL</EventType><EventDescription>Delivered</EventDescription><Address><City>Kandy</City><StateOrProvinceCode>ID</StateOrProvinceCode><PostalCode>90210</PostalCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></Address><ArrivalLocation>DELIVERY_LOCATION</ArrivalLocation></Events><Events><Timestamp>2018-01-03T10:10:35-07:00</Timestamp><EventType>OD</EventType><EventDescription>On FedEx vehicle for delivery</EventDescription><Address><City>Almost Kandy</City><StateOrProvinceCode>ID</StateOrProvinceCode><PostalCode>90211</PostalCode><CountryCode>US</CountryCode><CountryName>United States</CountryName><Residential>false</Residential></Address><ArrivalLocation>VEHICLE</ArrivalLocation></Events></TrackDetails></CompletedTrackDetails></TrackReply></SOAP-ENV:Body></SOAP-ENV:Envelope>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("POST", tt.u, responder)

			a := &FedEx{
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
						DateTime: parseDT("2018-01-03T11:12:45-07:00"),
						Location: Location{
							City: "Kandy", State: "ID", Country: "United States",
						},
						Status: "Delivered",
					},
					{
						DateTime: parseDT("2018-01-03T10:10:35-07:00"),
						Location: Location{
							City: "Almost Kandy", State: "ID", Country: "United States",
						},
						Status: "On FedEx vehicle for delivery",
					},
				},
				Expected: Expected{
					Delivery: "Delivered",
					Date:     time.Date(2018, 1, 3, 0, 0, 0, 0, time.UTC),
				},
				URL: fmt.Sprintf("https://www.fedex.com/apps/fedextrack/?action=track&tracknumbers=%v", strings.ToUpper(tt.name)),
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}

func parseDT(s string) time.Time {
	d, _ := time.Parse("2006-01-02T15:04:05-07:00", s)

	return d
}
