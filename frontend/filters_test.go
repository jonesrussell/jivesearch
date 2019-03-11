package frontend

import (
	"encoding/json"
	"html/template"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/congress"
	"github.com/jivesearch/jivesearch/instant/whois"
	"github.com/jivesearch/jivesearch/search"

	"github.com/jivesearch/jivesearch/instant/econ"
	"github.com/jivesearch/jivesearch/instant/econ/gdp"
	"github.com/jivesearch/jivesearch/instant/econ/population"

	"github.com/jivesearch/jivesearch/instant/currency"
	"github.com/jivesearch/jivesearch/instant/shortener"

	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"

	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

func TestAdd(t *testing.T) {
	type args struct {
		x int
		y int
	}

	for _, tt := range []struct {
		name string
		args
		want int
	}{
		{
			name: "1+1",
			args: args{1, 1},
			want: 2,
		},
		{
			name: "103+873",
			args: args{103, 873},
			want: 976,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := add(tt.args.x, tt.args.y)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestAnswerCSS(t *testing.T) {
	for _, tt := range []struct {
		name string
		want []string
	}{
		{
			name: "breach",
			want: []string{
				"http://127.0.0.1:8000/static/instant/owl.carousel.min.css",
				"http://127.0.0.1:8000/static/instant/breach/breach.css",
			},
		},
		{
			name: "calculator",
			want: []string{
				"http://127.0.0.1:8000/static/instant/calculator/calculator.css",
			},
		},
		{
			name: "whois",
			want: []string{},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			host := "http://127.0.0.1:8000"

			a := instant.Data{
				Type: instant.Type(tt.name),
			}

			got := answerCSS(host, a)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestAnswerJS(t *testing.T) {
	for _, tt := range []struct {
		name string
		want []string
	}{
		{
			name: "breach",
			want: []string{
				"http://127.0.0.1:8000/static/instant/owl.carousel.min.js",
				"http://127.0.0.1:8000/static/instant/breach/breach.js",
			},
		},
		{
			name: "calculator",
			want: []string{
				"http://127.0.0.1:8000/static/instant/calculator/calculator.js",
			},
		},
		{
			name: "whois",
			want: []string{},
		},
		{
			name: "wikidata nutrition",
			want: []string{
				"http://127.0.0.1:8000/static/instant/nutrition/nutrition.js",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			host := "http://127.0.0.1:8000"

			a := instant.Data{
				Type: instant.Type(tt.name),
			}

			got := answerJS(host, a)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestCommafy(t *testing.T) {
	for _, tt := range []struct {
		number interface{}
		want   string
	}{
		{
			number: 1000,
			want:   "1,000",
		},
		{
			number: 1023,
			want:   "1,023",
		},
		{
			number: -12000000,
			want:   "-12,000,000",
		},
		{
			number: int64(3743),
			want:   "3,743",
		},
		{
			number: -120,
			want:   "-120",
		},
		{
			number: 999000.01,
			want:   "999,000.01",
		},
		{
			number: 48915619813218,
			want:   "48,915,619,813,218",
		},
		{
			number: json.Number("12"),
			want:   "12",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := commafy(tt.number)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestHMACKey(t *testing.T) {
	type args struct {
		u      string
		secret string
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{"http://www.example.com/some/path/?query=string", "my_secret"},
			want: "LGSCFXg045ByB4ShdCHRIDlrPUDJ9eyFSrGz0HrtfAo=",
		},
		{
			name: "empty secret",
			args: args{"http://www.example.com/some/path/?query=string", ""},
			want: "oz13AtRiNq7h_rBVZMXccxPnDfnVHR12zd4honudDk4=",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hmacSecret = func() string { return tt.args.secret }

			got := hmacKey(tt.args.u)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestJSONMarshall(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  interface{}
	}{
		{
			"json object", `{"name":"bob"}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.arg)
			if err != nil {
				t.Fatal(err)
			}
			want := template.JS(b)

			got := jsonMarshal(tt.arg)
			if got != want {
				t.Fatalf("got %q; want %q", got, want)
			}
		})
	}
}

func TestJoin(t *testing.T) {
	for _, tt := range []struct {
		sl   []string
		want string
	}{
		{
			sl:   []string{"salt lake city", "utah", ""},
			want: "salt lake city, utah",
		},
		{
			sl:   []string{"boston", "MA", "US"},
			want: "boston, MA, US",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := join(tt.sl[0], tt.sl[1], tt.sl[2])
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestPercent(t *testing.T) {
	for _, tt := range []struct {
		number float64
		want   string
	}{
		{
			number: .2357,
			want:   "23.57%",
		},
		{
			number: .01527,
			want:   "1.53%",
		},
		{
			number: -1.0,
			want:   "-100.00%",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := percent(tt.number)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSafeHTML(t *testing.T) {
	for _, tt := range []struct {
		arg  string
		want template.HTML
	}{
		{
			arg:  "<!--[if lte IE 8]>",
			want: "<!--[if lte IE 8]>",
		},
		{
			arg:  "<!--[if gt IE 8]><!-->",
			want: "<!--[if gt IE 8]><!-->",
		},
		{
			arg:  "<script>some nasty javascript</script>",
			want: "<script>some nasty javascript</script>",
		},
	} {
		t.Run(tt.arg, func(t *testing.T) {
			got := safeHTML(tt.arg)

			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSortWHOISNameServers(t *testing.T) {
	for _, tt := range []struct {
		name string
		arg  []whois.NameServer
		want []whois.NameServer
	}{
		{
			name: "basic",
			arg: []whois.NameServer{
				{Name: "ns10.something.com"},
				{Name: "ns2.something.com"},
				{Name: "ns1.something.com"},
			},
			want: []whois.NameServer{
				{Name: "ns1.something.com"},
				{Name: "ns2.something.com"},
				{Name: "ns3.something.com"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := sortWHOISNameServers(tt.arg)

			if reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestStripHTML(t *testing.T) {
	for _, tt := range []struct {
		arg  string
		want string
	}{
		{
			arg:  `<a href="https://www.example.com">We want this!</a>`,
			want: "We want this!",
		},
		{
			arg:  `<div class="a bit" style="no style">this is the text we want</div>`,
			want: "this is the text we want",
		},
	} {
		t.Run(tt.arg, func(t *testing.T) {
			got := stripHTML(tt.arg)

			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSubtract(t *testing.T) {
	type args struct {
		x int
		y int
	}

	for _, tt := range []struct {
		name string
		args
		want int
	}{
		{
			name: "1-1",
			args: args{1, 1},
			want: 0,
		},
		{
			name: "103-10",
			args: args{103, 10},
			want: 93,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := subtract(tt.args.x, tt.args.y)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestSource(t *testing.T) {
	type args struct {
		src instant.Data
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "empty",
			args: args{instant.Data{}},
			want: "Jive Search",
		},
		{
			name: "breach",
			args: args{
				instant.Data{
					Type: "breach",
					Solution: &breach.Response{
						Provider: breach.HaveIBeenPwnedProvider,
					},
				},
			},
			want: `<br><img width="12" height="12" alt="Have I Been Pwned" src="/image/32x,s65582g6BoRK4EuPj871JNgrZwATF6AX6F7TF0uFh-F8=/https://haveibeenpwned.com/favicon.ico"/> <a href="https://haveibeenpwned.com/">Have I Been Pwned</a>`,
		},
		{
			name: "congress",
			args: args{
				instant.Data{
					Type: "congress",
					Solution: &congress.Response{
						Provider: congress.ProPublicaProvider,
					},
				},
			},
			want: `<br><img width="12" height="12" alt="ProPublica" src="/image/32x,s0OxD1aX5YGmf_Oi5xwt4iZ3ADaCio841_AVRWR0XZiI=/https://assets.propublica.org/prod/v3/images/favicon.ico"/> <a href="https://www.propublica.org/">ProPublica</a>`,
		},
		{
			name: "discography",
			args: args{
				instant.Data{
					Type: "discography",
				},
			},
			want: `<img width="12" height="12" alt="musicbrainz" src="/image/32x,sv4p1VZOkfT_gjscSjDjuToOCXgNXhcOxdBDjhYmwmsk=/https://musicbrainz.org/favicon.ico"/> <a href="https://musicbrainz.org/">MusicBrainz</a>`,
		},
		{
			name: "currency",
			args: args{
				instant.Data{
					Type: "currency",
					Solution: &instant.CurrencyResponse{
						Response: &currency.Response{
							CryptoProvider: currency.CryptoCompareProvider,
							ForexProvider:  currency.ECBProvider,
						},
					},
				},
			},
			want: `<img width="12" height="12" alt="European Central Bank" src="/image/32x,sojbRuJxSVjihgjhBCVOb63w6Xx3m8AdLx0eLr47VdA8=/http://www.ecb.europa.eu/favicon.ico"/> <a href="http://www.ecb.europa.eu/home/html/index.en.html">European Central Bank</a><br><img width="12" height="12" alt="CryptoCompare" src="/image/32x,stvpUZPbHHDno5wi-rZHX4YkppcMzE2yPC0FA2KyC4iM=/https://www.cryptocompare.com/media/20562/favicon.png?v=2"/> <a href="https://www.cryptocompare.com/">CryptoCompare</a>`,
		},
		{
			name: "fedex",
			args: args{
				instant.Data{
					Type: "fedex",
				},
			},
			want: `<img width="12" height="12" alt="fedex" src="/image/32x,sFXu9XPvd6hRjlea7BzoMkT0rEHPf0u7TawtAlUzQxvY=/http://www.fedex.com/favicon.ico"/> <a href="https://www.fedex.com">FedEx</a>`,
		},
		{
			name: "gdp",
			args: args{
				instant.Data{
					Type: "gdp",
					Solution: &instant.GDPResponse{
						Response: &gdp.Response{
							Provider: econ.TheWorldBankProvider,
						},
					},
				},
			},
			want: `<img width="12" height="12" alt="The World Bank" src="/image/32x,sr79IepQNuB0JCCgfeNKd5TpbGm4JSKlr9E4pUtiw9Ig=/https://www.worldbank.org/content/dam/wbr-redesign/logos/wbg-favicon.png"/> <a href="https://www.worldbank.org/">The World Bank</a>`,
		},
		{
			name: "population",
			args: args{
				instant.Data{
					Type: "population",
					Solution: &instant.PopulationResponse{
						Response: &population.Response{
							Provider: econ.TheWorldBankProvider,
						},
					},
				},
			},
			want: `<img width="12" height="12" alt="The World Bank" src="/image/32x,sr79IepQNuB0JCCgfeNKd5TpbGm4JSKlr9E4pUtiw9Ig=/https://www.worldbank.org/content/dam/wbr-redesign/logos/wbg-favicon.png"/> <a href="https://www.worldbank.org/">The World Bank</a>`,
		},
		{
			name: "stackoverflow",
			args: args{
				instant.Data{
					Type: "stackoverflow",
					Solution: &instant.StackOverflowAnswer{
						Answer: instant.SOAnswer{
							User: "bob",
						},
					},
				},
			},
			want: `bob via <img width="12" height="12" alt="stackoverflow" src="/image/32x,sT0tRYsDTt0J1npxPJ5N9YAHsrK7jWT0WcvRrCA0vRW8=/https://cdn.sstatic.net/Sites/stackoverflow/img/favicon.ico"/> <a href="https://stackoverflow.com/">Stack Overflow</a>`,
		},
		{
			name: "stock quote",
			args: args{
				instant.Data{
					Type: "stock quote",
					Solution: &stock.Quote{
						Provider: stock.IEXProvider,
					},
				},
			},
			want: `<img width="12" height="12" alt="IEX" src="/image/32x,sHbfM3QKtrjDw8v0skAKSmNQfZJ-C1OtMtjfBMNwsALI=/https://iextrading.com/favicon.ico"/> Data provided for free by <a href="https://iextrading.com/developer">IEX</a>.`,
		},
		{
			name: "url shortener",
			args: args{
				instant.Data{
					Type: "url shortener",
					Solution: &shortener.Response{
						Provider: shortener.IsGdProvider,
					},
				},
			},
			want: `<img width="12" height="12" alt="is.gd" src="/image/32x,sLbsPCKZT8roneQHXIKUMot4b2rdKGtVgiQ-bwvzLGMQ=/https://is.gd/isgd_favicon.ico"/> <a href="https://is.gd/">is.gd</a>`,
		},
		{
			name: "ups",
			args: args{
				instant.Data{
					Type: "ups",
				},
			},
			want: `<img width="12" height="12" alt="ups" src="/image/32x,s7SrucFg7UkzGUC-_R7m3DpE6fo5QP58vJgfJOx_0B7U=/https://www.ups.com/favicon.ico"/> <a href="https://www.ups.com">UPS</a>`,
		},
		{
			name: "usps",
			args: args{
				instant.Data{
					Type: "usps",
				},
			},
			want: `<img width="12" height="12" alt="usps" src="/image/32x,s2Yd8OZY8nuJHuzn8G36MHtysVdP4NyCIEI4g7M1gAOA=/https://www.usps.com/favicon.ico"/> <a href="https://www.usps.com">USPS</a>`,
		},
		{
			name: "weather",
			args: args{
				instant.Data{
					Type: "weather",
					Solution: &weather.Weather{
						Provider: weather.OpenWeatherMapProvider,
					},
				},
			},
			want: `<img width="12" height="12" alt="OpenWeatherMap" src="/image/32x,scZVABsAW_b194omixXeIrSTCiy1clQD-lTg3-51Lo84=/http://openweathermap.org/favicon.ico"/> <a href="http://openweathermap.org">OpenWeatherMap</a>`,
		},
		{
			name: "wikidata age",
			args: args{
				instant.Data{
					Type: "wikidata age",
				},
			},
			want: `<img width="12" height="12" alt="wikipedia" src="/image/32x,szl9NPdfHe0jt93aiLlox2zOB1DX2ThfpEHiI3AZWUpQ=/https://en.wikipedia.org/favicon.ico"/> <a href="https://www.wikipedia.org/">Wikipedia</a>`,
		},
		{
			name: "wikidata nutrition",
			args: args{
				instant.Data{
					Type: "wikidata nutrition",
				},
			},
			want: `<img width="12" height="12" alt="usda" src="/image/32x,s4sxklPbFqeSiQABvKk-LKPW1cG2gmPJJhy2Wehe73E8=/https://www.usda.gov/themes/usda/img/favicons/favicon.ico"/> <a href="https://www.usda.gov/">U.S. Department of Agriculture</a>`,
		},
		{
			name: "wikiquote",
			args: args{
				instant.Data{
					Type: "wikiquote",
				},
			},
			want: `<img width="12" height="12" alt="wikiquote" src="/image/32x,sybsABfe6inobFfifJrP1JfzqdReRgDujtUDZ6Kca5fA=/https://en.wikiquote.org/favicon.ico"/> <a href="https://www.wikiquote.org/">Wikiquote</a>`,
		},
		{
			name: "wiktionary",
			args: args{
				instant.Data{
					Type: "wiktionary",
				},
			},
			want: `<img width="12" height="12" alt="wiktionary" src="/image/32x,sFXYsXfk2L36AnwMAa75urF8XoD92aPCqQCo3eCvRSek=/https://www.wiktionary.org/static/favicon/piece.ico"/> <a href="https://www.wiktionary.org/">Wiktionary</a>`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := source(tt.args.src)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWeatherCode(t *testing.T) {
	for _, tt := range []struct {
		arg  weather.Description
		want string
	}{
		{
			weather.Clear, "icon-sun",
		},
		{
			weather.LightClouds, "icon-cloud-sun",
		},
		{
			weather.ScatteredClouds, "icon-cloud",
		},
		{
			weather.OvercastClouds, "icon-cloud-inv",
		},
		{
			weather.Extreme, "icon-cloud-flash-inv",
		},
		{
			weather.Rain, "icon-rain",
		},
		{
			weather.Snow, "icon-snowflake-o",
		},
		{
			weather.ThunderStorm, "icon-cloud-flash",
		},
		{
			weather.Windy, "icon-windy",
		},
		{
			"", "icon-sun",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := weatherCode(tt.arg)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestWeatherDailyForecast(t *testing.T) {
	for _, tt := range []struct {
		name      string
		timezone  string
		forecasts []*weather.Instant
		want      []*weatherDay
	}{
		{
			"empty", "",
			[]*weather.Instant{},
			[]*weatherDay{},
		},
		{
			"basic", "",
			[]*weather.Instant{
				{
					Date: time.Date(2018, 3, 15, 9, 0, 0, 0, time.UTC),
					Code: weather.LightClouds,
					Low:  47,
					High: 72,
				},
				{
					Date: time.Date(2018, 3, 15, 11, 0, 0, 0, time.UTC),
					Code: weather.OvercastClouds,
					Low:  44,
					High: 72,
				},
				{
					Date: time.Date(2018, 3, 15, 13, 0, 0, 0, time.UTC),
					Code: weather.LightClouds,
					Low:  48,
					High: 73,
				},
			},
			[]*weatherDay{
				{
					&weather.Instant{
						Date: time.Date(2018, 3, 15, 9, 0, 0, 0, time.UTC),
						Code: weather.LightClouds,
						Low:  44,
						High: 73,
					},
					"Thu 15",
					map[weather.Description]int{
						weather.LightClouds:    2,
						weather.OvercastClouds: 1,
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := weatherDailyForecast(tt.forecasts, tt.timezone)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestTitle(t *testing.T) {
	for _, tt := range []struct {
		s    interface{}
		want string
	}{
		{
			s:    "you should tItle tHIS strinG",
			want: "You Should TItle THIS StrinG",
		},
		{
			s:    search.Moderate,
			want: "Moderate",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := title(tt.s)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	for _, tt := range []struct {
		s    string
		len  int
		p    bool
		want string
	}{
		{
			s:    "This sentence should be truncated here and not go on and on and on and more on.",
			len:  39,
			p:    true,
			want: "This sentence should be truncated here ...",
		},
		{
			s:    "This sentence should be truncated here and not go on and on and on and more on.",
			len:  30,
			p:    false,
			want: "This sentence should be trunca...",
		},
		{
			s:    "This no truncate",
			len:  25,
			p:    true,
			want: "This no truncate",
		},
	} {
		t.Run(tt.want, func(t *testing.T) {
			got := truncate(tt.s, tt.len, tt.p)
			if got != tt.want {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

func TestWikiAmount(t *testing.T) {
	type args struct {
		quantity wikipedia.Quantity
		l        language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "meters",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11573"}, Amount: "2.16"},
				language.German,
			},
			want: "2.16 m",
		},
		{
			name: "cm",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q174728"}, Amount: "3"},
				language.French,
			},
			want: "3 cm",
		},
		{
			name: "inches",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q218593"}, Amount: "131"},
				language.French,
			},
			want: "333 cm",
		},
		{
			name: "kg",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"},
				language.Italian,
			},
			want: "147 kg",
		},
		{
			name: "meters (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11573"}, Amount: "2.16"},
				language.English,
			},
			want: `7'1"`,
		},
		{
			name: "cm  (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q174728"}, Amount: "3"},
				language.English,
			},
			want: `1.181103"`,
		},
		{
			name: "kg  (US)",
			args: args{
				wikipedia.Quantity{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"},
				language.English,
			},
			want: "324 lbs",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := tt.args.l.Region()

			got := wikiAmount(tt.args.quantity, r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiCanonical(t *testing.T) {
	type args struct {
		title string
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{"jimi hendrix was here"},
			want: "jimi_hendrix_was_here",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiCanonical(tt.args.title)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiData(t *testing.T) {
	type args struct {
		sol instant.Data
		l   language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "empty",
			args: args{
				instant.Data{
					Solution: []wikipedia.Quantity{},
				},
				language.English,
			},
			want: "",
		},
		{
			name: "kg",
			args: args{
				instant.Data{
					Solution: []wikipedia.Quantity{{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"}},
				},
				language.Italian,
			},
			want: "147 kg",
		},
		{
			name: "kg (cached version)",
			args: args{
				instant.Data{
					Solution: &[]wikipedia.Quantity{{Unit: wikipedia.Wikidata{ID: "Q11570"}, Amount: "147"}},
				},
				language.Italian,
			},
			want: "147 kg",
		},
		{
			name: "age (alive)",
			args: args{
				instant.Data{
					Solution: &instant.Age{
						Birthday: &instant.Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1972-12-31T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
				language.English,
			},
			want: `<em>Age:</em> 45 Years<br><span style="color:#666;">December 31, 1972</span>`,
		},
		{
			name: "age (at time of death)",
			args: args{
				instant.Data{
					Solution: &instant.Age{
						Birthday: &instant.Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1956-04-30T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: &instant.Death{
							Death: wikipedia.DateTime{
								Value:    "1984-03-13T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
				language.English,
			},
			want: `<em>Age at Death:</em> 27 Years<br><span style="color:#666;">April 30, 1956 - March 13, 1984</span>`,
		},
		{
			name: "birthday",
			args: args{
				instant.Data{
					Solution: &instant.Birthday{
						Birthday: wikipedia.DateTime{
							Value:    "2001-05-14T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
				language.English,
			},
			want: `May 14, 2001`,
		},
		{
			name: "death",
			args: args{
				instant.Data{
					Solution: &instant.Death{
						Death: wikipedia.DateTime{
							Value:    "2015-05-14T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
				language.English,
			},
			want: `May 14, 2015`,
		},
		{
			name: "unknown",
			args: args{
				instant.Data{Solution: 1}, language.English},
			want: "",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, _ := tt.args.l.Region()
			now = func() time.Time {
				return time.Date(2018, 02, 06, 20, 34, 58, 651387237, time.UTC)
			}

			got := wikiData(tt.args.sol, r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestWikidataClockTime(t *testing.T) {
	type args struct {
		t time.Time
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				time.Date(2016, 6, 5, 3, 2, 0, 0, time.UTC),
			},
			want: "3:02 AM",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikidataClockTime(tt.args.t)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikidataClockDate(t *testing.T) {
	type args struct {
		t time.Time
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				time.Date(2016, 6, 5, 3, 2, 0, 0, time.UTC),
			},
			want: "Sunday, June 5, 2016 (GMT+00)",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikidataClockDate(tt.args.t)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiDateTime(t *testing.T) {
	type args struct {
		dt wikipedia.DateTime
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "birthday",
			args: args{
				wikipedia.DateTime{
					Value:    "1972-12-31T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: "December 31, 1972",
		},
		{
			name: "year",
			args: args{
				wikipedia.DateTime{
					Value:    "1987",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: "1987",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiDateTime(tt.args.dt)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiJoin(t *testing.T) {
	type args struct {
		items     []wikipedia.Wikidata
		preferred []language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				[]wikipedia.Wikidata{
					{ID: "1", Labels: map[string]wikipedia.Text{
						"fr": {Text: "rock in french", Language: "fr"},
						"en": {Text: "rock", Language: "en"},
					}},
					{ID: "1", Labels: map[string]wikipedia.Text{
						"en": {Text: "rap", Language: "en"},
						"de": {Text: "rap in german", Language: "de"},
					}},
					{ID: "1", Labels: map[string]wikipedia.Text{
						"en": {Text: "country", Language: "en"},
					}},
				},
				[]language.Tag{
					language.English, language.French, language.German,
				},
			},
			want: "rock, rap, country",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiJoin(tt.args.items, tt.args.preferred)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikiLabel(t *testing.T) {
	type args struct {
		labels    map[string]wikipedia.Text
		preferred []language.Tag
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{
				map[string]wikipedia.Text{
					"en":    {Text: "english language", Language: "en"},
					"de":    {Text: "german language", Language: "de"},
					"sr-el": {Text: "this doesn't parse language", Language: "sr-el"},
				},
				[]language.Tag{
					language.English, language.French, language.German,
				},
			},
			want: "english language",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiLabel(tt.args.labels, tt.args.preferred)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestWikipediaItem(t *testing.T) {
	want := []*wikipedia.Item{{}}

	d := instant.Data{
		Solution: []*wikipedia.Item{{}},
	}

	got := wikipediaItem(d)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestWikiYears(t *testing.T) {
	type args struct {
		start interface{}
		end   interface{}
	}

	for _, tt := range []struct {
		name string
		args
		want int
	}{
		{
			name: "zero",
			args: args{
				time.Time{},
				time.Time{},
			},
			want: 0,
		},
		{
			name: "basic",
			args: args{
				time.Date(1975, 11, 17, 20, 34, 58, 651387237, time.UTC),
				time.Date(2017, 11, 18, 20, 34, 58, 651387237, time.UTC),
			},
			want: 42,
		},
		{
			name: "almost 42",
			args: args{
				time.Date(1975, 11, 17, 20, 34, 58, 651387237, time.UTC),
				time.Date(2017, 11, 16, 20, 34, 58, 651387237, time.UTC),
			},
			want: 41,
		},
		{
			name: "wikiDateTime",
			args: args{
				wikipedia.DateTime{
					Value:    "1854-12-31T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
				wikipedia.DateTime{
					Value:    "1912-04-30T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: 57,
		},
		{
			name: "wikiDateTime year",
			args: args{
				wikipedia.DateTime{
					Value:    "1794",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
				wikipedia.DateTime{
					Value:    "1954-02-14T00:00:00Z",
					Calendar: wikipedia.Wikidata{ID: "Q1985727"},
				},
			},
			want: 160,
		},
		{
			name: "wrong type",
			args: args{5, 12},
			want: 0,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := wikiYears(tt.args.start, tt.args.end)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
