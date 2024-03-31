package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jonesrussell/jivesearch/instant/parcel"
	"golang.org/x/text/language"
)

// UPSType is an answer Type
const UPSType Type = "ups"

// UPS is an instant answer
type UPS struct {
	parcel.Fetcher
	Answer
}

func (u *UPS) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *UPS) setUserAgent(r *http.Request) Answerer {
	return u
}

func (u *UPS) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *UPS) setType() Answerer {
	u.Type = UPSType
	return u
}

func (u *UPS) setRegex() Answerer {
	/*
		UPS Format:
		https://www.ups.com/ca/en/tracking/help/tracking/tnh.page
		1Z9999999999999999
		999999999999
		T9999999999
		999999999

		The InquiryNumber element is limited to a length of 9-34 digits as it can accept tracking numbers for:
			Small Package -- Either 18 digits (starting with 1Z) or 11 digits (starting with a single letter)
			Mail Innovations -- Typically 22 digits, starting with a 9
			Ground Freight -- Either 9 digits (all numbers) or 18 digits (starting with 1Z)
			Air Freight -- Either 10 digits (all numbers) or 18 digits (starting with 1Z)
			Ocean Freight -- Possibly 17 digits (per example in Dev Guide)
	*/

	// https://stackoverflow.com/questions/619977/regular-expression-patterns-for-tracking-numbers
	// There is also an undocumented checksum. Email from UPS would not provide me how to calculate it but a couple of hints:
	// https://github.com/jkeen/tracking_number_data/blob/70065359c64996e1537c46efc5d0638b24df105b/couriers/ups.json
	// (old) https://www.codeproject.com/Articles/21224/Calculating-the-UPS-Tracking-Number-Check-Digit
	// Probably only necessary if it triggers other IA's accidentally???
	// A Go library for checksum for UPS, FedEx, USPS: https://github.com/lensrentals/trackr

	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>(1Z ?[0-9A-Z]{3} ?[0-9A-Z]{3} ?[0-9A-Z]{2} ?[0-9A-Z]{4} ?[0-9A-Z]{3} ?[0-9A-Z]|T\d{3} ?\d{4} ?\d{3}|\d{22}))\b`))
	return u
}

func (u *UPS) solve(req *http.Request) Answerer {
	tn := strings.ToUpper(u.triggerWord)

	r, err := u.Fetch(tn)
	if err != nil {
		u.Err = err
		return u
	}

	u.Solution = r
	return u
}

func (u *UPS) tests() []test {
	testNumbers := []string{
		// Test numbers from "Tracking Web Service Developer Guide.pdf"
		// These won't return anything in the actual API but good for testing triggers
		"1z12345e0205271688",
		"1Z12345E6605272234",
		"1Z12345E0305271640",
		"1Z12345E0393657226",
		"1Z12345E1305277940",
		"1Z12345E6205277936",
		"1Z12345E1505270452",
		//"990728071",  // not working
		//"3251026119", // not working
		//"9102084383041101186729", // trips the FedEx regex...need to implement checksum
		//"cgish000116630", // not working
		"1Z648616E192760718",
		//"5548789114",        // not working
		//"ER751105042015062", // not working
		"1ZWX0692YP40636269",
	}

	tests := []test{}

	for _, n := range testNumbers {
		t := test{
			query: n,
			expected: []Data{
				{
					Type:      UPSType,
					Triggered: true,
					Solution: parcel.Response{
						TrackingNumber: strings.ToUpper(n),
						Updates: []parcel.Update{
							{
								DateTime: time.Date(2018, 3, 11, 2, 38, 0, 0, time.UTC),
								Location: parcel.Location{
									City: "Banahana", State: "ID", Country: "US",
								},
								Status: "Departure Scan",
							},
						},
						Expected: parcel.Expected{
							Delivery: "Scheduled Delivery",
							Date:     time.Date(2018, 3, 11, 0, 0, 0, 0, time.UTC),
						},
						URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", strings.ToUpper(n)),
					},
				},
			},
		}

		tests = append(tests, t)
	}

	return tests
}
