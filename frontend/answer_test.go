package frontend

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/status"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/whois"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

func TestAnswerHandler(t *testing.T) {
	bngs, err := bangsFromConfig()
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []struct {
		query string
		want  *response
	}{
		{
			"january birthstone",
			&response{
				status:   http.StatusOK,
				template: "jsonp",
				data: &AnswerResponse{
					HTML: `<div id=answer class=pure-u-1><div style=margin:15px;margin-bottom:5px>Garnet</div><div class=pure-u-1 style=margin-top:5px><div class=pure-u-1 style=margin-top:7px><div id=source class=pure-u-22-24 style=float:left;text-align:left;padding:15px><em>Source</em><br>Jive Search
<span style=float:right;text-align:right><a href=#open-widget onclick="document.getElementById('open-widget').style.display='block'">Get Widget</a></span></div></div></div></div>`,
					CSS:        []string{},
					JavaScript: []string{},
				},
			},
		},
		{
			"2+2",
			&response{
				status:   http.StatusOK,
				template: "jsonp",
				data: &AnswerResponse{
					HTML: `<div id=answer class=pure-u-1 style=width:323px;height:500px;padding:25px;padding-bottom:0><noscript><div id=answer class=pure-u-1><div style=margin:15px;margin-bottom:5px>4</div></div></noscript><div id=calculator style=display:none><div id=result tabindex=4>4</div><div id=main><div id=first-row><button id=clear class=del-bg>C</button>
<button class="btn-style operator opera-bg fall-back" value=%>%</button>
<button class="btn-style opera-bg align operator" value=/>/</button></div><div class=rows><button class="btn-style num-bg number first-child" value=7>7</button>
<button class="btn-style num-bg number" value=8>8</button>
<button class="btn-style num-bg number" value=9>9</button>
<button class="btn-style opera-bg operator" value=*>x</button></div><div class=rows><button class="btn-style num-bg number first-child" value=4>4</button>
<button class="btn-style num-bg number" value=5>5</button>
<button class="btn-style num-bg number" value=6>6</button>
<button class="btn-style opera-bg operator" value=->-</button></div><div class=rows><button class="btn-style num-bg number first-child" value=1>1</button>
<button class="btn-style num-bg number" value=2>2</button>
<button class="btn-style num-bg number" value=3>3</button>
<button class="btn-style opera-bg operator" value=+>+</button></div><div class=rows><button id=zero class="num-bg zero" value=0>0</button>
<button class="btn-style num-bg period fall-back" value=.>.</button>
<button id=eqn-bg class="eqn align" value="=">=</button></div></div></div><div class=pure-u-1 style=margin-top:5px><div class=pure-u-1 style=margin-top:7px><div id=source class=pure-u-22-24 style=float:left;text-align:left;padding:15px><em>Source</em><br>Jive Search
<span style=float:right;text-align:right><a href=#open-widget onclick="document.getElementById('open-widget').style.display='block'">Get Widget</a></span></div></div></div></div>`,
					CSS:        []string{"http://anything.com/static/instant/calculator/calculator.css"},
					JavaScript: []string{"http://anything.com/static/instant/calculator/calculator.js"},
				},
			},
		},
	} {
		t.Run(c.query, func(t *testing.T) {
			ParseTemplates()

			var matcher = language.NewMatcher(
				[]language.Tag{
					language.English,
					language.French,
				},
			)

			f := &Frontend{
				Brand: Brand{
					Host: "http://anything.com",
				},
				Bangs: bngs,
				Document: Document{
					Matcher: matcher,
				},
				Instant: &instant.Instant{
					BreachFetcher:        &mockBreachFetcher{},
					WikipediaFetcher:     &mockWikipediaFetcher{},
					StackOverflowFetcher: &mockStackOverflowFetcher{},
				},
				Wikipedia: Wikipedia{
					Matcher: matcher,
				},
			}

			f.Cache.Cacher = &mockCacher{}
			f.Cache.Instant = 10 * time.Second
			f.Cache.Search = 10 * time.Second

			req, err := http.NewRequest("GET", "/answer", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("q", c.query)
			req.URL.RawQuery = q.Encode()

			got := f.answerHandler(httptest.NewRecorder(), req)

			got.data.(*AnswerResponse).HTML, err = htmlMinify(got.data.(*AnswerResponse).HTML)
			if err != nil {
				t.Fatal(err)
			}

			/*
				fmt.Println(got.data.(*AnswerResponse).HTML)
				fmt.Println(c.want.data.(*AnswerResponse).HTML)
			*/

			if !reflect.DeepEqual(got, c.want) {
				//fmt.Println(got.data, c.want.data)
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}

func TestDetectType(t *testing.T) {
	for _, c := range []struct {
		name instant.Type
		want interface{}
	}{
		{instant.BirthStoneType, nil},
		{instant.BreachType, &breach.Response{}},
		{instant.CountryCodeType, &instant.CountryCodeResponse{}},
		{instant.CurrencyType, &instant.CurrencyResponse{}},
		{instant.DiscographyType, &[]discography.Album{}},
		{instant.FedExType, &parcel.Response{}},
		{instant.GDPType, &instant.GDPResponse{}},
		{instant.HashType, &instant.HashResponse{}},
		{instant.PopulationType, &instant.PopulationResponse{}},
		{instant.StackOverflowType, &instant.StackOverflowAnswer{}},
		{instant.StatusType, &status.Response{}},
		{instant.StockQuoteType, &stock.Quote{}},
		{instant.URLShortenerType, &shortener.Response{}},
		{instant.WeatherType, &weather.Weather{}},
		{instant.WHOISType, &whois.Response{}},
		{instant.WikipediaType, []*wikipedia.Item{}},
		{
			"wikidata age", &instant.Age{
				Birthday: &instant.Birthday{},
				Death:    &instant.Death{},
			},
		},
		{instant.WikidataBirthdayType, &instant.Birthday{}},
		{instant.WikidataDeathType, &instant.Death{}},

		{instant.WikidataHeightType, &[]wikipedia.Quantity{}},
		{instant.WikiquoteType, &[]string{}},
		{instant.WiktionaryType, &wikipedia.Wiktionary{}},
	} {
		t.Run(string(c.name), func(t *testing.T) {
			got := detectType(c.name)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}
