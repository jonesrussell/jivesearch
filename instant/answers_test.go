package instant

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/congress"
	curr "github.com/jivesearch/jivesearch/instant/currency"
	disc "github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/econ"
	ggdp "github.com/jivesearch/jivesearch/instant/econ/gdp"
	pop "github.com/jivesearch/jivesearch/instant/econ/population"
	"github.com/jivesearch/jivesearch/instant/nutrition"
	"github.com/jivesearch/jivesearch/instant/status"
	"github.com/jivesearch/jivesearch/instant/whois"

	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	so "github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

func answers(i Instant) []Answerer {
	return []Answerer{
		&BirthStone{},
		&Breach{Fetcher: i.BreachFetcher},
		&Calculator{},
		&CamelCase{},
		&Characters{},
		&Coin{},
		&Congress{Fetcher: i.CongressFetcher},
		&CountryCode{},
		&Discography{Fetcher: i.DiscographyFetcher},
		&DigitalStorage{},
		&FedEx{Fetcher: i.FedExFetcher},
		&Frequency{},
		&Currency{
			CryptoFetcher: i.CryptoFetcher,
			FXFetcher:     i.FXFetcher,
		},
		&GDP{GDPFetcher: i.GDPFetcher},
		&Hash{},
		&Speed{},
		&Length{},
		&Maps{LocationFetcher: i.LocationFetcher},
		&Minify{},
		&MortgageCalculator{},
		&Population{PopulationFetcher: i.PopulationFetcher},
		&Potus{},
		&Power{},
		&Prime{},
		&Random{},
		&Reverse{},
		&Shortener{Service: i.LinkShortener},
		&Stats{},
		&Status{Fetcher: i.StatusFetcher},
		&StockQuote{Fetcher: i.StockQuoteFetcher},
		&Temperature{},
		&USPS{Fetcher: i.USPSFetcher},
		&UPS{Fetcher: i.UPSFetcher},
		&URLDecode{},
		&URLEncode{},
		&UserAgent{},
		&StackOverflow{Fetcher: i.StackOverflowFetcher},
		&WHOIS{Fetcher: i.WHOISFetcher},
		&Weather{Fetcher: i.WeatherFetcher, LocationFetcher: i.LocationFetcher},
		&Wikipedia{
			LocationFetcher:  i.LocationFetcher,
			NutritionFetcher: i.NutritionFetcher,
			TimeZoneFetcher:  i.TimeZoneFetcher,
			Fetcher:          i.WikipediaFetcher,
		},
	}
}

// TestDetect runs the test cases for each instant answer.
func TestDetect(t *testing.T) {
	cases := []test{}

	i := Instant{
		QueryVar:        "q",
		BreachFetcher:   &mockBreachFetcher{},
		CongressFetcher: &mockCongressFetcher{},
		Currency: Currency{
			CryptoFetcher: &mockCryptoFetcher{},
			FXFetcher:     &mockFXFetcher{},
		},
		DiscographyFetcher:   &mockDiscographyFetcher{},
		FedExFetcher:         &mockFedExFetcher{},
		GDPFetcher:           &mockGDPFetcher{},
		LinkShortener:        &mockShortener{},
		LocationFetcher:      &mockLocationFetcher{},
		NutritionFetcher:     &mockNutritionFetcher{},
		PopulationFetcher:    &mockPopulationFetcher{},
		StackOverflowFetcher: &mockStackOverflowFetcher{},
		StatusFetcher:        &mockStatusFetcher{},
		StockQuoteFetcher:    &mockStockQuoteFetcher{},
		TimeZoneFetcher:      &mockTimeZoneFetcher{},
		UPSFetcher:           &mockUPSFetcher{},
		USPSFetcher:          &mockUSPSFetcher{},
		WeatherFetcher:       &mockWeatherFetcher{},
		WHOISFetcher:         &mockWHOISFetcher{},
		WikipediaFetcher:     &mockWikipediaFetcher{},
	}

	for j, ia := range answers(i) {
		if len(ia.tests()) == 0 {
			t.Fatalf("No tests for answer #%d", j)
		}
		cases = append(cases, ia.tests()...)
	}

	for _, c := range cases {
		t.Run(c.query, func(t *testing.T) {
			ctx := fmt.Sprintf(`(query: %q, user agent: %q)`, c.query, c.userAgent)

			v := url.Values{}
			v.Set("q", c.query)

			r := &http.Request{
				Form:   v,
				Header: make(http.Header),
			}

			r.Header.Set("User-Agent", c.userAgent)
			r.Header.Set("X-Forwarded-For", c.ip.String())

			for _, ia := range answers(i) {
				if triggered := i.Trigger(ia, r, language.English); triggered {
					now = func() time.Time {
						return time.Date(2016, 6, 5, 3, 2, 0, 0, time.UTC)
					}

					got := i.Solve(ia, r)
					if got.Err != nil {
						continue
					}

					var solved bool

					for _, expected := range c.expected {
						if reflect.DeepEqual(got, expected) {
							solved = true
							break
						}
					}
					if !solved {
						fmt.Println(got.Type)
						t.Errorf("Instant answer failed %v", ctx)
						t.Errorf("got %+v;", got)
						//t.Errorf("got %+v;", got.Solution.(*Clock).Time)
						t.Errorf("want ")
						for _, expected := range c.expected {
							t.Errorf("    %+v\n", expected)
							//t.Errorf("    %+v\n", expected.Solution.(*Clock).Time)
						}
						t.FailNow()
					}
					break
				}
			}
		})
	}
}

func TestGetIPAddress(t *testing.T) {
	type args struct {
		remoteAddr    string
		xRealIP       string
		xForwardedFor []string
	}

	for _, tt := range []struct {
		name string
		args
		want net.IP
	}{
		{
			"no header",
			args{remoteAddr: "161.59.224.138"},
			net.ParseIP("161.59.224.138"),
		},
		{
			"x-real-ip",
			args{xRealIP: "161.59.224.139"},
			net.ParseIP("161.59.224.139"),
		},
		{
			"x-forwarded-for",
			args{xForwardedFor: []string{"161.59.224.140"}},
			net.ParseIP("161.59.224.140"),
		},
		{
			"remote addr and x-forwarded-for",
			args{remoteAddr: "161.59.224.138", xForwardedFor: []string{"161.59.224.140"}},
			net.ParseIP("161.59.224.140"),
		},
		{
			"multiple x-forwarded-for",
			args{xForwardedFor: []string{"161.59.224.140", "161.59.224.141"}},
			net.ParseIP("161.59.224.140"),
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := http.Header{}

			h.Set("X-Real-IP", tt.args.xRealIP)
			for _, address := range tt.args.xForwardedFor {
				h.Add("X-Forwarded-For", address)
			}

			r := &http.Request{
				RemoteAddr: tt.args.remoteAddr,
				Header:     h,
			}

			got := getIPAddress(r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}

		})
	}
}

// mock FedEx Fetcher
type mockFedExFetcher struct{}

func (f *mockFedExFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
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
		URL: fmt.Sprintf("https://www.fedex.com/apps/fedextrack/?action=track&tracknumbers=%v", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

type mockCongressFetcher struct{}

func (m *mockCongressFetcher) FetchMembers(location *congress.Location) (*congress.Response, error) {
	return &congress.Response{
		Location: &congress.Location{
			Short: "UT",
			State: "Utah",
		},
		Role: congress.House,
		Members: []congress.Member{
			{
				Name:         "Rob Bishop",
				District:     1,
				Gender:       "M",
				Party:        "R",
				Twitter:      "RepRobBishop",
				Facebook:     "RepRobBishop",
				NextElection: 2018,
			},
			{
				Name:         "Chris Stewart",
				District:     2,
				Gender:       "M",
				Party:        "R",
				Twitter:      "RepChrisStewart",
				Facebook:     "RepChrisStewart",
				NextElection: 2018,
			},
			{
				Name:         "John Curtis",
				District:     3,
				Gender:       "M",
				Party:        "R",
				Twitter:      "RepJohnCurtis",
				Facebook:     "",
				NextElection: 2018,
			},
			{
				Name:         "Mia Love",
				District:     4,
				Gender:       "F",
				Party:        "R",
				Twitter:      "repmialove",
				Facebook:     "",
				NextElection: 2018,
			},
		},
		Provider: congress.ProPublicaProvider,
	}, nil
}

func (m *mockCongressFetcher) FetchSenators(location *congress.Location) (*congress.Response, error) {
	return &congress.Response{
		Location: location,
		Role:     congress.Senators,
		Members: []congress.Member{
			{
				Name:         "Orrin G. Hatch",
				Gender:       "M",
				Party:        "R",
				Twitter:      "SenOrrinHatch",
				Facebook:     "senatororrinhatch",
				NextElection: 2018,
			},
			{
				Name:         "Mike Lee",
				Gender:       "M",
				Party:        "R",
				Twitter:      "SenMikeLee",
				Facebook:     "senatormikelee",
				NextElection: 2022,
			},
		},
		Provider: congress.ProPublicaProvider,
	}, nil
}

type mockCryptoFetcher struct{}

func (m *mockCryptoFetcher) Fetch() (*curr.Response, error) {
	return &curr.Response{
		Base: curr.USD,
		History: map[string][]*curr.Rate{
			curr.BTC.Short: {
				{
					DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
					Rate:     1.12,
				},
				{
					DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
					Rate:     1.1,
				},
			},
			curr.LTC.Short: {
				{
					DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
					Rate:     1.5,
				},
				{
					DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
					Rate:     1.6,
				},
			},
		},
		CryptoProvider: curr.CryptoCompareProvider,
	}, nil
}

type mockFXFetcher struct{}

func (m *mockFXFetcher) Fetch() (*curr.Response, error) {
	return &curr.Response{
		Base: curr.USD,
		History: map[string][]*curr.Rate{
			curr.JPY.Short: {
				{
					DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
					Rate:     1.12,
				},
				{
					DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
					Rate:     1.1,
				},
			},
			curr.GBP.Short: {
				{
					DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
					Rate:     1.5,
				},
				{
					DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
					Rate:     1.6,
				},
			},
		},
		ForexProvider: curr.ECBProvider,
	}, nil
}

// mock gdp fetcher
type mockGDPFetcher struct{}

func (m *mockGDPFetcher) Fetch(country string, start time.Time, end time.Time) (*ggdp.Response, error) {
	return &ggdp.Response{
		History: []ggdp.Instant{
			{
				Date:  time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 4,
			},
			{
				Date:  time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 2,
			},
			{
				Date:  time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 18,
			},
		},
		Provider: econ.TheWorldBankProvider,
	}, nil
}

// mock location fetcher
type mockLocationFetcher struct{}

func (l *mockLocationFetcher) Fetch(ip net.IP) (*location.City, error) {
	c := &location.City{}
	c.City.Names = map[string]string{"en": "Someville"}
	c.Subdivisions = []location.Subdivision{
		{Names: map[string]string{"en": "SomeState"}},
	}
	c.Country.Names = map[string]string{"en": "SomeCountry"}
	c.Location.Latitude = 12
	c.Location.Longitude = 18
	return c, nil
}

// mock nutrition
type mockNutritionFetcher struct{}

func (m *mockNutritionFetcher) Fetch(ndbnos []string) (*nutrition.Response, error) {
	var r *nutrition.Response

	if reflect.DeepEqual(ndbnos, []string{"01123"}) {
		r = &nutrition.Response{
			Provider: "Mock Response",
			Foods: []nutrition.Food{
				{
					Name:        "Egg, whole, raw, fresh",
					FoodGroup:   "Dairy and Egg Products",
					Corporation: "",
					Nutrients: []nutrition.Nutrient{
						{
							ID:    "212",
							Name:  "Sodium",
							Unit:  "mg",
							Value: json.Number("12"),
							Measures: []nutrition.Measure{
								{
									Label:      "large",
									Equivalent: 50,
									Units:      "g",
									Quantity:   1,
									Value:      json.Number("72"),
								},
								{
									Label:      "extra large",
									Equivalent: 56,
									Units:      "g",
									Quantity:   1,
									Value:      json.Number("80"),
								},
							},
						},
					},
				},
			},
		}
	} else if reflect.DeepEqual(ndbnos, []string{"12"}) {
		r = &nutrition.Response{
			Provider: "Mock Response",
			Foods: []nutrition.Food{
				{
					Name:        "Big Mac",
					FoodGroup:   "Some Category",
					Corporation: "McDowell's",
					Nutrients: []nutrition.Nutrient{
						{
							ID:    "208",
							Name:  "Energy",
							Unit:  "kcal",
							Value: json.Number("554"),
							Measures: []nutrition.Measure{
								{
									Label:      "1 size",
									Equivalent: 12,
									Units:      "g",
									Quantity:   1,
									Value:      json.Number("720"),
								},
							},
						},
					},
				},
			},
		}
	}

	return r, nil
}

func (m *mockNutritionFetcher) Lookup(query string) ([]*nutrition.ItemResponse, error) {
	var r []*nutrition.ItemResponse

	switch query {
	case "egg":
		r = []*nutrition.ItemResponse{
			{
				Name:  "egg",
				NDBNO: "01123",
			},
		}
	case "big mac":
		r = []*nutrition.ItemResponse{
			{
				Name:         "Big Mac",
				NDBNO:        "12",
				Manufacturer: "McDowell's",
			},
		}
	}

	return r, nil
}

type mockPopulationFetcher struct{}

func (m *mockPopulationFetcher) Fetch(country string, start time.Time, end time.Time) (*pop.Response, error) {
	return &pop.Response{
		History: []pop.Instant{
			{
				Date:  time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 4,
			},
			{
				Date:  time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 2,
			},
			{
				Date:  time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC),
				Value: 18,
			},
		},
		Provider: econ.TheWorldBankProvider,
	}, nil
}

// mock Stack Overflow Fetcher
type mockStackOverflowFetcher struct{}

func (s *mockStackOverflowFetcher) Fetch(query string, tags []string) (so.Response, error) {
	resp := so.Response{}

	switch query {
	case "loop":
		if reflect.DeepEqual(tags, []string{"php"}) {
			resp = so.Response{
				Items: []so.Item{
					{
						Answers: []so.Answer{
							{
								Owner: so.Owner{
									DisplayName: "NikiC",
								},
								Score: 1273,
								Body:  "an answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/10057671/how-does-php-foreach-actually-work",
						Title: "How does PHP &#39;foreach&#39; actually work?",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"c++"}) {
			resp = so.Response{
				Items: []so.Item{
					{
						Answers: []so.Answer{
							{
								Owner: so.Owner{
									DisplayName: "JamesT",
								},
								Score: 90210,
								Body:  "a very good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/c++-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"go"}) {
			resp = so.Response{
				Items: []so.Item{
					{
						Answers: []so.Answer{
							{
								Owner: so.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/go-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"macos"}) {
			resp = so.Response{
				Items: []so.Item{
					{
						Answers: []so.Answer{
							{
								Owner: so.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/macos-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		} else if reflect.DeepEqual(tags, []string{"regex"}) {
			resp = so.Response{
				Items: []so.Item{
					{
						Answers: []so.Answer{
							{
								Owner: so.Owner{
									DisplayName: "Danny Zuko",
								},
								Score: 90210,
								Body:  "a superbly good answer",
							},
						},
						Link:  "https://stackoverflow.com/questions/90210/regex-loop",
						Title: "Some made-up question",
					},
				},
				QuotaMax:       300,
				QuotaRemaining: 197,
			}
		}

	default:
	}

	return resp, nil
}

// mock stock quote Fetcher
type mockStockQuoteFetcher struct{}

func (s *mockStockQuoteFetcher) Fetch(ticker string) (*stock.Quote, error) {
	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return nil, err
	}

	q := &stock.Quote{}

	switch ticker {
	case "AAPL":
		q.Ticker = "AAPL"
		q.Name = "Apple Inc."
		q.Exchange = stock.NASDAQ
	case "BRK.A":
		q.Ticker = "BRK.A"
		q.Name = "Berkshire Hathaway"
		q.Exchange = stock.NYSE
	}

	q.Last = stock.Last{
		Price:         171.42,
		Time:          time.Unix(1522090355062/1000, 0).In(location),
		Change:        6.48,
		ChangePercent: 0.03929,
	}
	q.History = []stock.EOD{
		{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
		{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
	}
	q.Provider = stock.IEXProvider

	return q, nil
}

// mock UPS Fetcher
type mockUPSFetcher struct{}

func (u *mockUPSFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
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
		URL: fmt.Sprintf("https://wwwapps.ups.com/WebTracking/processInputRequest?AgreeToTermsAndConditions=yes&InquiryNumber1=%v&TypeOfInquiryNumber=T&error_carried=true&loc=en-us&sort_by=status&tracknums_displayed=1", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

// mock USPS Fetcher
type mockUSPSFetcher struct{}

func (u *mockUSPSFetcher) Fetch(trackingNumber string) (parcel.Response, error) {
	r := parcel.Response{
		TrackingNumber: strings.ToUpper(trackingNumber),
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
		URL: fmt.Sprintf("https://tools.usps.com/go/TrackConfirmAction?origTrackNum=%v", strings.ToUpper(trackingNumber)),
	}

	return r, nil
}

// mock weather Fetcher
type mockWeatherFetcher struct {
	location.Fetcher
}

func (m *mockWeatherFetcher) FetchByCity(city string) (*weather.Weather, error) {
	w := &weather.Weather{
		City: "Bogota",
		Current: &weather.Instant{
			Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
			Code:        weather.ScatteredClouds,
			Temperature: 59,
			Low:         55,
			High:        63,
			Wind:        4.7,
			Clouds:      40,
			Rain:        0,
			Snow:        0,
			Pressure:    1014,
			Humidity:    33,
		},
		Forecast: []*weather.Instant{
			{
				Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 97,
				Low:         84,
				High:        97,
				Wind:        3.94,
				Pressure:    888.01,
				Humidity:    14,
			},
			{
				Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 95,
				Low:         85,
				High:        99,
				Wind:        10.76,
				Pressure:    886.87,
				Humidity:    13,
			},
		},
		Provider: weather.OpenWeatherMapProvider,
	}

	return w, nil
}

func (m *mockWeatherFetcher) FetchByLatLong(lat, long float64, timezone string) (*weather.Weather, error) {
	w := &weather.Weather{
		City: "Bountiful",
		Current: &weather.Instant{
			Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
			Code:        weather.ScatteredClouds,
			Temperature: 59,
			Low:         55,
			High:        63,
			Wind:        4.7,
			Clouds:      40,
			Rain:        0,
			Snow:        0,
			Pressure:    1014,
			Humidity:    33,
		},
		Forecast: []*weather.Instant{
			{
				Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 97,
				Low:         84,
				High:        97,
				Wind:        3.94,
				Pressure:    888.01,
				Humidity:    14,
			},
			{
				Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 95,
				Low:         85,
				High:        95,
				Wind:        10.76,
				Pressure:    886.87,
				Humidity:    13,
			},
		},
		Provider: weather.OpenWeatherMapProvider,
		TimeZone: "America/Denver",
	}

	return w, nil
}

func (m *mockWeatherFetcher) FetchByZip(zip int) (*weather.Weather, error) {
	w := &weather.Weather{
		City: "Bountiful",
		Current: &weather.Instant{
			Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
			Code:        weather.ScatteredClouds,
			Temperature: 59,
			Low:         55,
			High:        63,
			Wind:        4.7,
			Clouds:      40,
			Rain:        0,
			Snow:        0,
			Pressure:    1014,
			Humidity:    33,
		},
		Forecast: []*weather.Instant{
			{
				Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 97,
				Low:         84,
				High:        97,
				Wind:        3.94,
				Pressure:    888.01,
				Humidity:    14,
			},
			{
				Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
				Code:        weather.Clear,
				Temperature: 95,
				Low:         85,
				High:        95,
				Wind:        10.76,
				Pressure:    886.87,
				Humidity:    13,
			},
		},
		Provider: weather.OpenWeatherMapProvider,
	}

	return w, nil
}

// mock WHOIS Fetcher
type mockWHOISFetcher struct{}

func (f *mockWHOISFetcher) Fetch(domain string) (*whois.Response, error) {
	resp := &whois.Response{}

	switch domain {
	case "google.com":
		resp = &whois.Response{
			Domain:   domain,
			DomainID: "2138514_DOMAIN_COM-VRSN",
			Status:   "registered",
			Nameservers: []whois.NameServer{
				{Name: "ns3.google.com"}, {Name: "ns4.google.com"}, {Name: "ns1.google.com"}, {Name: "ns2.google.com"},
			},
			Available:  false,
			Registered: true,
		}
	}

	return resp, nil
}

// mock Wikipedia Fetcher
type mockWikipediaFetcher struct{}

func (mf *mockWikipediaFetcher) Fetch(query string, lang language.Tag) ([]*wikipedia.Item, error) {
	switch query {
	case "bob marley":
		return []*wikipedia.Item{
			{
				Wikidata: &wikipedia.Wikidata{
					Claims: &wikipedia.Claims{
						Birthday: []wikipedia.DateTime{
							{
								Value:    "1945-02-06T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: []wikipedia.DateTime{
							{
								Value:    "1981-05-11T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
			},
		}, nil
	case "jimi hendrix":
		return []*wikipedia.Item{
			{
				Wikidata: &wikipedia.Wikidata{
					Claims: &wikipedia.Claims{
						Birthday: []wikipedia.DateTime{
							{
								Value:    "1942-11-27T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: []wikipedia.DateTime{
							{
								Value:    "1970-09-18T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
			},
		}, nil

	case "shaquille o'neal":
		return []*wikipedia.Item{
			{
				Wikidata: &wikipedia.Wikidata{
					Claims: &wikipedia.Claims{
						Height: []wikipedia.Quantity{
							{
								Amount: "2.16",
								Unit:   wikipedia.Wikidata{ID: "Q11573"},
							},
						},
						Weight: []wikipedia.Quantity{
							{
								Amount: "147",
								Unit:   wikipedia.Wikidata{ID: "Q11573"},
							},
						},
					},
				},
			},
		}, nil
	case "michael jordan":
		return []*wikipedia.Item{
			{
				Wikiquote: wikipedia.Wikiquote{
					Quotes: []string{
						"I can accept failure. Everyone fails at something. But I can't accept not trying (no hard work)",
						"ball is life",
					},
				},
			},
		}, nil
	case "guitar":
		return []*wikipedia.Item{
			{
				Wiktionary: wikipedia.Wiktionary{
					Title: "guitar",
					Definitions: []*wikipedia.Definition{
						{Part: "noun", Meaning: "musical instrument"},
					},
				},
			},
		}, nil
	case "sydney":
		return []*wikipedia.Item{
			{
				Wikidata: &wikipedia.Wikidata{
					Claims: &wikipedia.Claims{
						Coordinate: []wikipedia.Coordinate{
							{
								Latitude:  []float64{-33.8667},
								Longitude: []float64{151.2},
							},
						},
					},
				},
			},
		}, nil
	default:
		return []*wikipedia.Item{
			{
				Wikidata: &wikipedia.Wikidata{
					Claims: &wikipedia.Claims{
						Coordinate: []wikipedia.Coordinate{
							{
								Latitude:  []float64{123.45},
								Longitude: []float64{56.789},
							},
						},
					},
				},
			},
		}, nil
	}
}

func (mf *mockWikipediaFetcher) Setup() error {
	return nil
}

type mockBreachFetcher struct{}

func (m *mockBreachFetcher) Fetch(account string) (*breach.Response, error) {
	r := &breach.Response{
		Account: "test@example.com",
		Breaches: []breach.Breach{
			{
				Name:        "000webhost",
				Domain:      "000webhost.com",
				Date:        time.Date(2015, 3, 1, 0, 0, 0, 0, time.UTC),
				Count:       14936670,
				Description: "Some description here.",
				Items:       []string{"Email addresses", "IP addresses", "Names", "Passwords"},
			},
			{
				Name:        "8tracks",
				Domain:      "8tracks.com",
				Date:        time.Date(2017, 6, 27, 0, 0, 0, 0, time.UTC),
				Count:       7990619,
				Description: "Another description here.",
				Items:       []string{"Email addresses", "Passwords"},
			},
		},
		Provider: breach.HaveIBeenPwnedProvider,
	}

	return r, nil
}

type mockDiscographyFetcher struct{}

func (m *mockDiscographyFetcher) Fetch(artist string) ([]disc.Album, error) {
	u, _ := url.Parse("http://coverartarchive.org/release/1/2-250..jpg")
	return []disc.Album{
		{
			Name:      "Are You Experienced",
			Published: time.Date(1970, 9, 18, 0, 0, 0, 0, time.UTC),
			Image: disc.Image{
				URL: u,
			},
		},
	}, nil
}

type mockShortener struct{}

func (m *mockShortener) Shorten(u *url.URL) (*shortener.Response, error) {
	shrt, _ := url.Parse("http://shrt.url")

	return &shortener.Response{
		Original: u,
		Short:    shrt,
		Provider: "mockShortener",
	}, nil
}

type mockStatusFetcher struct{}

func (m *mockStatusFetcher) Fetch(domain string) (*status.Response, error) {
	r := &status.Response{}

	switch domain {
	case "example.com":
		r = &status.Response{
			Domain:   "example.com",
			Port:     80,
			Status:   1,
			IP:       "1.2.3.4",
			Code:     200,
			Time:     1.2,
			Provider: status.IsItUpProvider,
		}
	case "something.com":
		r = &status.Response{
			Domain:   "something.com",
			Port:     80,
			Status:   1,
			IP:       "1.2.3.5",
			Code:     301,
			Time:     .23,
			Provider: status.IsItUpProvider,
		}
	}

	return r, nil
}

type mockTimeZoneFetcher struct{}

func (m *mockTimeZoneFetcher) Fetch(lat, lon float64) (string, error) {
	if lat == -33.8667 && lon == 151.2 {
		return "Australia/Sydney", nil
	}

	return "America/Denver", nil
}
