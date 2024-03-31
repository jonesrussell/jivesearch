package parcel

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/jonesrussell/jivesearch/log"
)

// UPS holds settings for the UPS API
type UPS struct {
	HTTPClient *http.Client
	User       string
	Password   string
	Key        string
}

// UPSResponse is UPS's raw JSON response
type UPSResponse struct {
	Response
	TrackResponse struct {
		Response struct {
			ResponseStatus struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"ResponseStatus"`
			TransactionReference struct {
				CustomerContext string `json:"CustomerContext"`
			} `json:"TransactionReference"`
		} `json:"Response"`
		Shipment struct {
			InquiryNumber struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
				Value       string `json:"Value"`
			} `json:"InquiryNumber"`
			ShipperNumber   string `json:"ShipperNumber"`
			ShipmentAddress []struct {
				Type struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Type"`
				Address struct {
					AddressLine       interface{} `json:"AddressLine"`
					City              string      `json:"City"`
					StateProvinceCode string      `json:"StateProvinceCode"`
					PostalCode        string      `json:"PostalCode"`
					CountryCode       string      `json:"CountryCode"`
				} `json:"Address"`
			} `json:"ShipmentAddress"`
			ShipmentWeight struct {
				UnitOfMeasurement struct {
					Code string `json:"Code"`
				} `json:"UnitOfMeasurement"`
				Weight string `json:"Weight"`
			} `json:"ShipmentWeight"`
			Service struct {
				Code        string `json:"Code"`
				Description string `json:"Description"`
			} `json:"Service"`
			ReferenceNumber interface{} `json:"ReferenceNumber"`
			DeliveryDetail  struct {
				Type struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Type"`
				Date string `json:"Date"`
			} `json:"DeliveryDetail"`
			Package struct {
				TrackingNumber       string `json:"TrackingNumber"`
				PackageServiceOption struct {
					Type struct {
						Code        string `json:"Code"`
						Description string `json:"Description"`
					} `json:"Type"`
				} `json:"PackageServiceOption"`
				Activity []struct {
					ActivityLocation struct {
						Address struct {
							City              string `json:"City"`
							StateProvinceCode string `json:"StateProvinceCode"`
							CountryCode       string `json:"CountryCode"`
						} `json:"Address"`
					} `json:"ActivityLocation"`
					Status struct {
						Type        string `json:"Type"`
						Description string `json:"Description"`
						Code        string `json:"Code"`
					} `json:"Status"`
					Date string `json:"Date"`
					Time string `json:"Time"`
				} `json:"Activity"`
				Message struct {
					Code        string `json:"Code"`
					Description string `json:"Description"`
				} `json:"Message"`
				PackageWeight struct {
					UnitOfMeasurement struct {
						Code string `json:"Code"`
					} `json:"UnitOfMeasurement"`
					Weight string `json:"Weight"`
				} `json:"PackageWeight"`
				ReferenceNumber []struct {
					Code  string `json:"Code"`
					Value string `json:"Value"`
				} `json:"ReferenceNumber"`
			} `json:"Package"`
		} `json:"Shipment"`
		Disclaimer string `json:"Disclaimer"`
	} `json:"TrackResponse"`
}

// UnmarshalJSON sets the Response fields
func (r *UPSResponse) UnmarshalJSON(b []byte) error {
	type alias UPSResponse
	raw := &alias{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	r.Response.TrackingNumber = raw.TrackResponse.Shipment.Package.TrackingNumber

	for _, a := range raw.TrackResponse.Shipment.Package.Activity {
		d := a.Date + a.Time
		dt, err := time.Parse("20060102150405", d)
		if err != nil {
			log.Debug.Println(err) // this isn't serious enough to warrant a return
		}

		up := Update{
			DateTime: dt,
			Status:   a.Status.Description,
		}

		up.Location.City = a.ActivityLocation.Address.City
		up.Location.State = a.ActivityLocation.Address.StateProvinceCode
		up.Location.Country = a.ActivityLocation.Address.CountryCode
		r.Response.Updates = append(r.Response.Updates, up)
	}

	r.Response.Expected.Delivery = raw.TrackResponse.Shipment.DeliveryDetail.Type.Description
	if raw.TrackResponse.Shipment.DeliveryDetail.Date != "" {
		r.Response.Expected.Date, err = time.Parse("20060102", raw.TrackResponse.Shipment.DeliveryDetail.Date)
		if err != nil {
			return err
		}
	}

	return r.buildURL()
}

func (u *UPS) buildRequest(number string) (*http.Request, error) {
	// test url := "https://wwwcie.ups.com/rest/Track"
	url := "https://onlinetools.ups.com/rest/Track"

	jsonStr := []byte(fmt.Sprintf(`{
		"UPSSecurity": {
			"UsernameToken": {
				"Username": "%v",
				"Password": "%v"
			},
			"ServiceAccessToken": {
				"AccessLicenseNumber": "%v"
			}
		},
		"TrackRequest": {
			"Request": {
				"RequestOption": "1",
				"TransactionReference": {
					"CustomerContext": "Your Test Case Summary Description"
				}
			},
			"InquiryNumber": "%v"
		}
	}`, u.User, u.Password, u.Key, number))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}

	// Docs say the following headers are required but API seems to work w/out them
	req.Header.Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	req.Header.Set("Access-Control-Allow-Methods", "POST")
	req.Header.Set("Access-Control-Allow-Origin", "*")
	req.Header.Set("Content-Type", "application/json")
	return req, err
}

// Fetch retrieves from the UPS API
func (u *UPS) Fetch(trackingNumber string) (Response, error) {
	r := UPSResponse{}

	req, err := u.buildRequest(trackingNumber)
	if err != nil {
		return r.Response, err
	}

	resp, err := u.HTTPClient.Do(req)
	if err != nil {
		return r.Response, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.Response, err
}

func (r *UPSResponse) buildURL() error {
	u, err := url.Parse("https://wwwapps.ups.com/WebTracking/processInputRequest")
	if err != nil {
		return err
	}

	q := u.Query()
	q.Set("sort_by", "status")
	q.Set("error_carried", "true")
	q.Set("tracknums_displayed", "1")
	q.Set("TypeOfInquiryNumber", "T")
	q.Set("loc", "en-us")
	q.Set("InquiryNumber1", r.Response.TrackingNumber)
	q.Set("AgreeToTermsAndConditions", "yes")
	u.RawQuery = q.Encode()
	r.Response.URL = u.String()
	return nil
}
