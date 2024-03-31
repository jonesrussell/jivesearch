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

// FedExType is an answer Type
const FedExType Type = "fedex"

// FedEx is an instant answer
type FedEx struct {
	parcel.Fetcher
	Answer
}

func (f *FedEx) setQuery(r *http.Request, qv string) Answerer {
	f.Answer.setQuery(r, qv)
	return f
}

func (f *FedEx) setUserAgent(r *http.Request) Answerer {
	return f
}

func (f *FedEx) setLanguage(lang language.Tag) Answerer {
	f.language = lang
	return f
}

func (f *FedEx) setType() Answerer {
	f.Type = FedExType
	return f
}

func (f *FedEx) setRegex() Answerer {
	// https://stackoverflow.com/questions/619977/regular-expression-patterns-for-tracking-numbers
	// https://github.com/jkeen/tracking_number_data/blob/70065359c64996e1537c46efc5d0638b24df105b/couriers/fedex.json
	// Probably only necessary if it triggers other IA's accidentally???
	f.regex = append(f.regex, regexp.MustCompile(`(?i)\b(?P<trigger>[0-9]{10}|[0-9]{12}|[0-9]{15}|[0-9]{20})\b`))

	return f
}

func (f *FedEx) solve(req *http.Request) Answerer {
	r, err := f.Fetch(f.triggerWord)
	if err != nil {
		f.Err = err
		return f
	}

	f.Solution = r
	return f
}

func (f *FedEx) tests() []test {
	testNumbers := []string{
		// Express and Ground
		"449044304137821",
		"149331877648230",
		"020207021381215",
		"403934084723025",
		"920241085725456",
		"568838414941",
		"039813852990618",
		"231300687629630",
		"797806677146",
		"377101283611590",
		"852426136339213",
		"797615467620",
		"957794015041323",
		"076288115212522",
		"581190049992",
		"122816215025810",
		"843119172384577",
		"070358180009382",
		// SmartPost
		"02394653001023698293",
		"61292701078443410536",
		"61292700726653585070",
		"02394653018047202719",
		// LTL
		"2873008051",
		"1960003216",
		"1208673524",
		"1636374036",
	}

	tests := []test{}

	for _, n := range testNumbers {
		t := test{
			query: n,
			expected: []Data{
				{
					Type:      FedExType,
					Triggered: true,
					Solution: parcel.Response{
						TrackingNumber: strings.ToUpper(n),
						Updates: []parcel.Update{
							{
								DateTime: time.Date(2018, 1, 3, 11, 12, 45, 0, time.Local),
								Location: parcel.Location{
									City: "Kandy", State: "ID", Country: "United States",
								},
								Status: "Delivered",
							},
							{
								DateTime: time.Date(2018, 1, 3, 10, 10, 35, 0, time.Local),
								Location: parcel.Location{
									City: "Almost Kandy", State: "ID", Country: "United States",
								},
								Status: "On FedEx vehicle for delivery",
							},
						},
						Expected: parcel.Expected{
							Delivery: "Delivered",
							Date:     time.Date(2018, 1, 3, 0, 0, 0, 0, time.UTC),
						},
						URL: fmt.Sprintf("https://www.fedex.com/apps/fedextrack/?action=track&tracknumbers=%v", strings.ToUpper(n)),
					},
				},
			},
		}

		tests = append(tests, t)
	}

	return tests
}
