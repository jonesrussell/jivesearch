package parcel

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jonesrussell/jivesearch/log"
)

// USPS holds settings for the USPS API
type USPS struct {
	HTTPClient *http.Client
	User       string
	Password   string
}

// USPSResponse is USPS's raw XML response
type USPSResponse struct {
	Response
	uspsTrackResponse
}

type uspsTrackResponse struct {
	TrackInfo *struct {
		AttrID       string       `xml:"ID,attr"`
		TrackDetail  []*uspsEvent `xml:"TrackDetail,omitempty"`
		TrackSummary *uspsEvent   `xml:"TrackSummary,omitempty"`
	} `xml:"TrackInfo,omitempty"`
}

type uspsEvent struct {
	AuthorizedAgent *struct {
		Text string `xml:",chardata"`
	} `xml:"AuthorizedAgent,omitempty"`
	Event *struct {
		Text string `xml:",chardata"`
	} `xml:"Event,omitempty"`
	EventCity *struct {
		Text string `xml:",chardata"`
	} `xml:"EventCity,omitempty"`
	EventCountry *struct {
		Text string `xml:",chardata"`
	} `xml:"EventCountry,omitempty"`
	EventDate *struct {
		Text string `xml:",chardata"`
	} `xml:"EventDate,omitempty"`
	EventState *struct {
		Text string `xml:",chardata"`
	} `xml:"EventState,omitempty"`
	EventTime *struct {
		Text string `xml:",chardata"`
	} `xml:"EventTime,omitempty"`
	EventZIPCode *struct {
		Text string `xml:",chardata"`
	} `xml:"EventZIPCode,omitempty"`
	FirmName *struct{} `xml:"FirmName,omitempty"`
	Name     *struct{} `xml:"Name,omitempty"`
}

// UnmarshalXML sets the Response fields
func (r *USPSResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	parseEvent := func(e *uspsEvent) Update {
		d := fmt.Sprintf("%v %v", e.EventDate.Text, e.EventTime.Text)
		dt, err := time.Parse("January 2, 2006 15:04 pm", d)
		if err != nil {
			log.Debug.Println(err) // this isn't serious enough to warrant a return
		}

		up := Update{
			DateTime: dt,
			Status:   e.Event.Text,
		}

		up.Location.City = e.EventCity.Text
		up.Location.State = e.EventState.Text
		up.Location.Country = e.EventCountry.Text
		return up
	}

	type alias USPSResponse
	raw := &alias{}

	err := d.DecodeElement(&raw, &start)
	if err != nil {
		return err
	}

	r.Response.TrackingNumber = raw.uspsTrackResponse.TrackInfo.AttrID

	up := parseEvent(raw.uspsTrackResponse.TrackInfo.TrackSummary)
	r.Response.Updates = append(r.Response.Updates, up)

	for _, a := range raw.uspsTrackResponse.TrackInfo.TrackDetail {
		up := parseEvent(a)
		r.Response.Updates = append(r.Response.Updates, up)
	}

	r.Response.URL = fmt.Sprintf("https://tools.usps.com/go/TrackConfirmAction?origTrackNum=%v", r.Response.TrackingNumber)
	return err
}

// Fetch retrieves from the USPS API
func (u *USPS) Fetch(trackingNumber string) (Response, error) {
	r := USPSResponse{}

	uu, err := url.Parse("http://production.shippingapis.com/ShippingAPI.dll")
	if err != nil {
		return r.Response, err
	}

	// Adding <Revision>1</Revision> will give a bit more detail but nothing useful
	x := fmt.Sprintf(`<TrackFieldRequest USERID="%v"><TrackID ID="%v"></TrackID></TrackFieldRequest>`,
		u.User, trackingNumber,
	)

	q := uu.Query()
	q.Set("API", "TrackV2")
	q.Set("XML", x)
	uu.RawQuery = q.Encode()

	resp, err := http.Get(uu.String())
	if err != nil {
		return r.Response, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&r)
	return r.Response, err
}
