package parcel

import (
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/log"
)

// FedEx holds settings for the FedEx API
type FedEx struct {
	HTTPClient *http.Client
	Key        string
	Password   string
	Account    string
	Meter      string
}

// FedExResponse is FedEx's raw XML response
type FedExResponse struct {
	Response
	Body struct {
		TrackReply struct {
			HighestSeverity string `xml:"HighestSeverity"`
			Notifications   []struct {
				Severity         string
				Source           string
				Code             int
				Message          string
				LocalizedMessage string
			}
			Version struct {
				ServiceID    string
				Major        int
				Intermediate int
				Minor        int
			}
			CompletedTrackDetails struct {
				HighestSeverity string
				Notifications   []struct {
					Severity         string
					Source           string
					Code             int
					Message          string
					LocalizedMessage string
				}
				DuplicateWaybill bool
				MoreData         bool
				TrackDetails     struct {
					TrackingNumber                 string
					TrackingNumberUniqueIdentifier string
					Notification                   struct {
						Severity         string
						Source           string
						Code             int
						Message          string
						LocalizedMessage string
					}
					StatusDetail struct {
						CreationTime string
						Code         string
						Description  string
						Location     struct {
							StreetLines         string
							City                string
							StateOrProvinceCode string
							CountryCode         string
							CountryName         string
							Residential         bool
						}
						AncillaryDetails []struct {
							Reason            string
							ReasonDescription string
						}
					}
					CarrierCode                          string
					OperatingCompanyOrCarrierDescription string
					OtherIdentifiers                     []struct {
						PackageIdentifier struct {
							Type  string
							Value string
						}
					}
					Service struct {
						Type             string
						Description      string
						ShortDescription string
					}
					PackageWeight struct {
						Units string
						Value float64
					}
					ShipmentWeight struct {
						Units string
						Value float64
					}
					Packaging             string
					PackagingType         string
					PackageSequenceNumber int
					PackageCount          int
					SpecialHandlings      []struct {
						Type        string
						Description string
						PaymentType string
					}
					ShipTimestamp           string
					ActualDeliveryTimestamp string
					DestinationAddress      struct {
						StreetLines         string
						City                string
						StateOrProvinceCode string
						CountryCode         string
						CountryName         string
						Residential         bool
					}
					ActualDeliveryAddress struct {
						StreetLines         string
						City                string
						StateOrProvinceCode string
						CountryCode         string
						CountryName         string
						Residential         bool
					}
					DeliveryLocationType                   string
					DeliveryLocationDescription            string
					DeliveryAttempts                       int
					DeliverySignatureName                  string
					TotalUniqueAddressCountInConsolidation int
					NotificationEventsAvailable            string
					RedirectToHoldEligibility              string
					Events                                 []struct {
						Timestamp                  string
						EventType                  string
						EventDescription           string
						StatusExceptionCode        string
						StatusExceptionDescription string
						Address                    struct {
							StreetLines         string
							City                string
							StateOrProvinceCode string
							CountryCode         string
							CountryName         string
							Residential         bool
						}
						ArrivalLocation string
					}
				}
			}
		} `xml:"TrackReply"`
	} `xml:"Body"`
}

// UnmarshalXML sets the Response fields
func (r *FedExResponse) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	type alias FedExResponse
	raw := &alias{}

	err := d.DecodeElement(&raw, &start)
	if err != nil {
		return err
	}

	r.Response.TrackingNumber = raw.Body.TrackReply.CompletedTrackDetails.TrackDetails.TrackingNumber

	for _, e := range raw.Body.TrackReply.CompletedTrackDetails.TrackDetails.Events {
		dt, err := time.Parse("2006-01-02T15:04:05-07:00", e.Timestamp)
		if err != nil {
			log.Debug.Println(err) // this isn't serious enough to warrant a return
		}

		up := Update{
			DateTime: dt,
			Status:   e.EventDescription,
		}

		up.Location.City = e.Address.City
		up.Location.State = e.Address.StateOrProvinceCode
		up.Location.Country = e.Address.CountryName
		r.Response.Updates = append(r.Response.Updates, up)
	}

	r.Response.Expected.Delivery = raw.Body.TrackReply.CompletedTrackDetails.TrackDetails.StatusDetail.Description
	dt := raw.Body.TrackReply.CompletedTrackDetails.TrackDetails.StatusDetail.CreationTime
	if dt != "" {
		r.Response.Expected.Date, err = time.Parse("2006-01-02T15:04:05", dt)
		if err != nil {
			return err
		}
	}

	r.Response.URL = fmt.Sprintf("https://www.fedex.com/apps/fedextrack/?action=track&tracknumbers=%v", r.Response.TrackingNumber)

	return err
}

func (f *FedEx) buildRequest(number string) string {
	/*
		Elements left out from their example in docs:
		<v14:Destination>
			<v14:GeographicCoordinates>rates evertitque	aequora</v14:GeographicCoordinates>
		</v14:Destination>

	*/
	return fmt.Sprintf(`<?xml version="1.0"?>
		<soapenv:Envelope xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/" xmlns:v14="http://fedex.com/ws/track/v14">
		<soapenv:Header/>
		<soapenv:Body>
			<v14:TrackRequest>
				<v14:WebAuthenticationDetail>
					<v14:ParentCredential>
						<v14:Key>%v</v14:Key>
						<v14:Password>%v</v14:Password>
					</v14:ParentCredential>
					<v14:UserCredential>
						<v14:Key>%v</v14:Key>
						<v14:Password>%v</v14:Password>
					</v14:UserCredential>
				</v14:WebAuthenticationDetail>
				<v14:ClientDetail>
					<v14:AccountNumber>%v</v14:AccountNumber>
					<v14:MeterNumber>%v</v14:MeterNumber>
				</v14:ClientDetail>
				<v14:TransactionDetail>
					<v14:CustomerTransactionId>Track By Number_v14</v14:CustomerTransactionId>
					<v14:Localization>
						<v14:LanguageCode>EN</v14:LanguageCode>
						<v14:LocaleCode>US</v14:LocaleCode>
					</v14:Localization>
				</v14:TransactionDetail>
				<v14:Version>
					<v14:ServiceId>trck</v14:ServiceId>
					<v14:Major>14</v14:Major>
					<v14:Intermediate>0</v14:Intermediate>
					<v14:Minor>0</v14:Minor>
				</v14:Version>
				<v14:SelectionDetails>
					<v14:PackageIdentifier>
						<v14:Type>TRACKING_NUMBER_OR_DOORTAG</v14:Type>
						<v14:Value>%v</v14:Value>
					</v14:PackageIdentifier>
					<v14:ShipmentAccountNumber/>
					<v14:SecureSpodAccount/>
				</v14:SelectionDetails>
				<v14:ProcessingOptions>INCLUDE_DETAILED_SCANS</v14:ProcessingOptions>
			</v14:TrackRequest>
		</soapenv:Body>
		</soapenv:Envelope>`,
		f.Key, f.Password, f.Key, f.Password, f.Account, f.Meter, number,
	)
}

// Fetch retrieves from the FedEx API
func (f *FedEx) Fetch(trackingNumber string) (Response, error) {
	r := FedExResponse{}

	url := "https://ws.fedex.com/web-services"
	x := f.buildRequest(trackingNumber)

	resp, err := http.Post(url, "text/xml", strings.NewReader(x))
	if err != nil {
		return r.Response, err
	}
	defer resp.Body.Close()

	err = xml.NewDecoder(resp.Body).Decode(&r)
	return r.Response, err
}
