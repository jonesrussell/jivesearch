package frontend

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/congress"
	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/status"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/whois"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/log"
	"golang.org/x/text/language"
)

// AnswerResponse is an instant answer response
type AnswerResponse struct {
	HTML       string   `json:"html"`
	CSS        []string `json:"css"`
	JavaScript []string `json:"javascript"`
}

func (f *Frontend) answerHandler(w http.ResponseWriter, r *http.Request) *response {
	d := f.getData(r)

	resp := &response{
		status:   http.StatusOK,
		template: "jsonp",
		err:      nil,
	}

	ic := make(chan instant.Data)
	go f.getAnswer(r, d, ic)

	d.Instant = <-ic
	resp.data = d

	a := &AnswerResponse{}

	// get the html string
	var buf bytes.Buffer
	answerTemplate := "answer"
	tmpl, ok := templates[answerTemplate]
	if !ok {
		resp.status = http.StatusInternalServerError
		resp.err = fmt.Errorf("template doesn't exist: %q", answerTemplate)
		return resp
	}

	if err := tmpl.Execute(&buf, resp.data); err != nil {
		resp.status, resp.err = http.StatusInternalServerError, err
		return resp
	}

	a.HTML = buf.String()
	a.CSS = answerCSS(f.Brand.Host, d.Instant)
	a.JavaScript = answerJS(f.Brand.Host, d.Instant)
	resp.data = a

	return resp
}

func (f *Frontend) getAnswer(r *http.Request, dd data, ic chan instant.Data) {
	lang, _, _ := f.Wikipedia.Matcher.Match(dd.Context.Preferred...)
	key := cacheKey("instant", lang, f.detectRegion(lang, r), r.URL)

	v, err := f.Cache.Get(key)
	if err != nil {
		log.Info.Println(err)
	}

	if v != nil {
		ir := &Instant{
			instant.Data{},
		}

		if err := json.Unmarshal(v.([]byte), &ir); err != nil {
			log.Info.Println(err)
		}

		ic <- ir.Data
		return
	}

	// only need to trigger the maps instant answer if maps or images nav selected
	var onlyMaps bool
	if dd.Context.T == "maps" || dd.Context.T == "images" {
		onlyMaps = true
	}

	var d = f.Cache.Instant

	res := f.DetectInstantAnswer(r, lang, onlyMaps)

	var cache bool

	switch res.Type {
	case instant.WikidataClockType, instant.CoinTossType, instant.LocalWeatherType, instant.RandomType, instant.UserAgentType: // only local weather
		cache = false
	case instant.CurrencyType, instant.StockQuoteType, instant.FedExType, instant.UPSType, instant.USPSType:
		d = 1 * time.Minute
		cache = true
	case instant.WikipediaType: // I can't figure out how to cache this without errors...
		cache = false
	default:
		cache = true
	}

	if cache {
		if d > f.Cache.Instant {
			d = f.Cache.Instant
		}

		if err := f.Cache.Put(key, res, d); err != nil {
			log.Info.Println(err)
		}
	}

	ic <- res
}

// DetectInstantAnswer triggers the instant answers
func (f *Frontend) DetectInstantAnswer(r *http.Request, lang language.Tag, onlyMaps bool) instant.Data {
	var answers []instant.Answerer

	// select all answers by default, unless user chooses maps
	switch onlyMaps {
	case true:
		answers = []instant.Answerer{
			&instant.Maps{LocationFetcher: f.Instant.LocationFetcher},
			&instant.Wikipedia{
				Fetcher: f.Instant.WikipediaFetcher,
			},
		}
	default:
		answers = []instant.Answerer{
			&instant.BirthStone{},
			&instant.Breach{
				Fetcher: f.Instant.BreachFetcher,
			},
			&instant.Calculator{},
			&instant.CamelCase{},
			&instant.Characters{},
			&instant.Coin{},
			&instant.Congress{
				Fetcher: f.Instant.CongressFetcher,
			},
			&instant.CountryCode{},
			&instant.Currency{
				CryptoFetcher: f.Instant.CryptoFetcher,
				FXFetcher:     f.Instant.FXFetcher,
			},
			&instant.Discography{Fetcher: f.Instant.DiscographyFetcher},
			&instant.DigitalStorage{},
			&instant.FedEx{Fetcher: f.Instant.FedExFetcher},
			&instant.Frequency{},
			&instant.GDP{GDPFetcher: f.Instant.GDPFetcher},
			&instant.Hash{},
			&instant.Speed{}, // trigger "miles per hour" b/f "miles"
			&instant.Length{},
			&instant.Maps{LocationFetcher: f.Instant.LocationFetcher},
			&instant.Minify{},
			&instant.MortgageCalculator{},
			&instant.Population{PopulationFetcher: f.Instant.PopulationFetcher},
			&instant.Potus{},
			&instant.Power{},
			&instant.Prime{},
			&instant.Random{},
			&instant.Reverse{},
			&instant.Shortener{Service: f.Instant.LinkShortener},
			&instant.Stats{},
			&instant.Status{Fetcher: f.Instant.StatusFetcher},
			&instant.StockQuote{Fetcher: f.Instant.StockQuoteFetcher},
			&instant.Temperature{},
			&instant.USPS{Fetcher: f.Instant.USPSFetcher},
			&instant.UPS{Fetcher: f.Instant.UPSFetcher},
			&instant.URLDecode{},
			&instant.URLEncode{},
			&instant.UserAgent{},
			&instant.StackOverflow{Fetcher: f.Instant.StackOverflowFetcher},
			&instant.Weather{Fetcher: f.Instant.WeatherFetcher, LocationFetcher: f.Instant.LocationFetcher},
			&instant.WHOIS{Fetcher: f.Instant.WHOISFetcher},
			&instant.Wikipedia{
				LocationFetcher: f.Instant.LocationFetcher,
				TimeZoneFetcher: f.Instant.TimeZoneFetcher,
				Fetcher:         f.Instant.WikipediaFetcher,
			}, // always keep this last so that Wikipedia Box will trigger if none other
		}
	}

	// Necessary to use goroutines??? setSolution called only when triggered.
	// Also, the order of some answers matters, like Wikipedia, which is a catch-all
	for _, ia := range answers {
		if triggered := f.Instant.Trigger(ia, r, lang); triggered {
			sol := f.Instant.Solve(ia, r)
			if sol.Err != nil {
				log.Debug.Println(sol.Err)
				continue
			}

			return sol
		}
	}

	return instant.Data{}
}

// UnmarshalJSON unmarshals an instant answer to the correct data structure
func (d *Instant) UnmarshalJSON(b []byte) error {
	type alias Instant
	raw := &alias{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	j, err := json.Marshal(raw.Solution)
	if err != nil {
		return err
	}

	d.Data = raw.Data
	d.Solution = raw.Solution

	s := detectType(raw.Type)
	if s == nil { // a string
		return nil
	}

	d.Solution = s
	return json.Unmarshal(j, d.Solution)
}

// detectType returns the proper data structure for an instant answer type
func detectType(t instant.Type) interface{} {
	var v interface{}

	switch t {
	case instant.BreachType:
		v = &breach.Response{}
	case instant.CongressType:
		v = &congress.Response{}
	case instant.CountryCodeType:
		v = &instant.CountryCodeResponse{}
	case instant.DiscographyType:
		v = &[]discography.Album{}
	case instant.CurrencyType:
		v = &instant.CurrencyResponse{}
	case instant.FedExType, instant.UPSType, instant.USPSType:
		v = &parcel.Response{}
	case instant.GDPType:
		v = &instant.GDPResponse{}
	case instant.HashType:
		v = &instant.HashResponse{}
	case instant.PopulationType:
		v = &instant.PopulationResponse{}
	case instant.StackOverflowType:
		v = &instant.StackOverflowAnswer{}
	case instant.StatusType:
		v = &status.Response{}
	case instant.StockQuoteType:
		v = &stock.Quote{}
	case instant.URLShortenerType:
		v = &shortener.Response{}
	case instant.LocalWeatherType, instant.WeatherType:
		v = &weather.Weather{}
	case instant.WHOISType:
		v = &whois.Response{}
	case instant.WikipediaType:
		v = []*wikipedia.Item{}
	case instant.WikidataAgeType:
		v = &instant.Age{
			Birthday: &instant.Birthday{},
			Death:    &instant.Death{},
		}
	case instant.WikidataBirthdayType:
		v = &instant.Birthday{}
	case instant.WikidataClockType:
		v = &instant.Clock{}
	case instant.WikidataDeathType:
		v = &instant.Death{}
	case instant.WikidataHeightType, instant.WikidataWeightType:
		v = &[]wikipedia.Quantity{}
	case instant.WikiquoteType:
		v = &[]string{}
	case instant.WiktionaryType:
		v = &wikipedia.Wiktionary{}
	default: // a string
		return nil
	}

	return v
}
