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

// USPSType is an answer Type
const USPSType Type = "usps"

// USPS is an instant answer
type USPS struct {
	parcel.Fetcher
	Answer
}

func (u *USPS) setQuery(r *http.Request, qv string) Answerer {
	u.Answer.setQuery(r, qv)
	return u
}

func (u *USPS) setUserAgent(r *http.Request) Answerer {
	return u
}

func (u *USPS) setLanguage(lang language.Tag) Answerer {
	u.language = lang
	return u
}

func (u *USPS) setType() Answerer {
	u.Type = USPSType
	return u
}

func (u *USPS) setRegex() Answerer {
	// https://stackoverflow.com/questions/619977/regular-expression-patterns-for-tracking-numbers
	// A Go library for checksum for UPS, FedEx, USPS: https://github.com/lensrentals/trackr
	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>\d{30}\b)|(\b91\d+\b)|(\b\d{20})\b`))
	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>E\D{1}\d{9}\D{2}$|^9\d{15,21})\b`))
	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>91[0-9]+)\b`))
	u.regex = append(u.regex, regexp.MustCompile(`(?i)\b(?P<trigger>[A-Za-z]{2}[0-9]+US)\b`))
	return u
}

func (u *USPS) solve(req *http.Request) Answerer {
	tn := strings.ToUpper(u.triggerWord)

	r, err := u.Fetch(tn)
	if err != nil {
		u.Err = err
		return u
	}

	u.Solution = r
	return u
}

func (u *USPS) tests() []test {
	testNumbers := []string{
		//"70160910000108310009", // certified...trips fedex
		//"23153630000057728970",   // signature confirmation...trips fedex
		"RE360192014US",          // registered mail
		"eL595811950US",          // priority express
		"9374889692090270407075", // regular
	}

	tests := []test{}

	for _, n := range testNumbers {
		t := test{
			query: n,
			expected: []Data{
				{
					Type:      USPSType,
					Triggered: true,
					Solution: parcel.Response{
						TrackingNumber: strings.ToUpper(n),
						Updates: []parcel.Update{
							{
								DateTime: time.Date(2018, 3, 12, 13, 57, 0, 0, time.UTC),
								Location: parcel.Location{
									City: "Some City", State: "ID", Country: "",
								},
								Status: "Delivered",
							},
							{
								DateTime: time.Date(2018, 3, 14, 8, 13, 0, 0, time.UTC),
								Location: parcel.Location{
									City: "Close to Some City", State: "ID", Country: "",
								},
								Status: "Out for Delivery",
							},
							{
								DateTime: time.Date(2018, 3, 14, 7, 11, 0, 0, time.UTC),
								Location: parcel.Location{
									City: "Almost", State: "ID", Country: "",
								},
								Status: "Almost there dude",
							},
						},
						URL: fmt.Sprintf("https://tools.usps.com/go/TrackConfirmAction?origTrackNum=%v", strings.ToUpper(n)),
					},
				},
			},
		}

		tests = append(tests, t)
	}

	return tests
}
