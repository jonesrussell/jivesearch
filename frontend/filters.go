package frontend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/congress"
	"github.com/jivesearch/jivesearch/instant/whois"
	"github.com/jivesearch/jivesearch/search"

	humanize "github.com/dustin/go-humanize"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/instant/currency"
	"github.com/jivesearch/jivesearch/instant/econ"
	"github.com/jivesearch/jivesearch/instant/shortener"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"github.com/jivesearch/jivesearch/log"
	"golang.org/x/text/language"
)

var funcMap = template.FuncMap{
	"Add":                  add,
	"AnswerCSS":            answerCSS,
	"AnswerJS":             answerJS,
	"Commafy":              commafy,
	"HMACKey":              hmacKey,
	"Join":                 join,
	"JSONMarshal":          jsonMarshal,
	"Now":                  now,
	"Percent":              percent,
	"SafeHTML":             safeHTML,
	"Source":               source,
	"SortWHOISNameServers": sortWHOISNameServers,
	"StripHTML":            stripHTML,
	"Subtract":             subtract,
	"Title":                title,
	"Truncate":             truncate,
	"WeatherCode":          weatherCode,
	"WeatherDailyForecast": weatherDailyForecast,
	"WikiAmount":           wikiAmount,
	"WikiCanonical":        wikiCanonical,
	"WikidataClockTime":    wikidataClockTime,
	"WikidataClockDate":    wikidataClockDate,
	"WikiData":             wikiData,
	"WikiDateTime":         wikiDateTime,
	"WikiJoin":             wikiJoin,
	"WikiLabel":            wikiLabel,
	"WikipediaItem":        wikipediaItem,
	"WikiYears":            wikiYears,
}

func add(x, y int) int {
	return x + y
}

var addStaticPrefix = func(host, f string) string {
	return fmt.Sprintf("%v/static/instant/%v", host, f)
}

func answerCSS(host string, a instant.Data) []string {
	files := []string{}

	switch a.Type {
	case "breach":
		files = []string{
			addStaticPrefix(host, "owl.carousel.min.css"),
			addStaticPrefix(host, "breach/breach.css"),
		}
	case "calculator":
		files = []string{
			addStaticPrefix(host, "calculator/calculator.css"),
		}
	case "currency":
		files = []string{
			addStaticPrefix(host, "currency/currency.css"),
		}
	case "discography":
		files = []string{
			addStaticPrefix(host, "owl.carousel.min.css"),
			addStaticPrefix(host, "discography/discography.css"),
		}
	case "gdp":
		files = []string{
			addStaticPrefix(host, "gdp/gdp.css"),
		}
	case "maps":
		files = []string{
			addStaticPrefix(host, "maps/mapbox.css"),
			addStaticPrefix(host, "maps/mapbox_directions.css"),
		}
	case "population":
		files = []string{
			addStaticPrefix(host, "population/population.css"),
		}
	case "stock quote":
		files = []string{
			addStaticPrefix(host, "stock_quotes/stock_quotes.css"),
		}
	case "unit converter":
		files = []string{
			addStaticPrefix(host, "unit_converter/unit_converter.css"),
		}
	case "local weather", "weather":
		files = []string{
			addStaticPrefix(host, "weather/weather.css"),
		}
	}

	return files
}

func answerJS(host string, a instant.Data) []string {
	files := []string{}

	switch a.Type {
	case "breach":
		files = []string{
			addStaticPrefix(host, "owl.carousel.min.js"),
			addStaticPrefix(host, "breach/breach.js"),
		}
	case "calculator":
		files = []string{
			addStaticPrefix(host, "calculator/calculator.js"),
		}
	case "currency":
		files = []string{
			addStaticPrefix(host, "d3.v4.min.js"),
			addStaticPrefix(host, "currency/currency.js"),
		}
	case "discography":
		files = []string{
			addStaticPrefix(host, "owl.carousel.min.js"),
			addStaticPrefix(host, "discography/discography.js"),
		}
	case "gdp":
		files = []string{
			addStaticPrefix(host, "d3.v4.min.js"),
			addStaticPrefix(host, "gdp/gdp.js"),
		}
	case "maps":
		files = []string{
			addStaticPrefix(host, "maps/mapbox.js"),
			addStaticPrefix(host, "maps/mapbox_directions.js"),
		}
	case "mortgage calculator":
		files = []string{
			addStaticPrefix(host, "mortgage_calculator/mortgage_calculator.js"),
		}
	case "population":
		files = []string{
			addStaticPrefix(host, "d3.v4.min.js"),
			addStaticPrefix(host, "population/population.js"),
		}
	case "stock quote":
		files = []string{
			addStaticPrefix(host, "d3.v4.min.js"),
			addStaticPrefix(host, "stock_quotes/stock_quotes.js"),
		}
	case "unit converter":
		files = []string{
			addStaticPrefix(host, "unit_converter/unit_converter.js"),
		}
	case "minify":
		/*
					prettydiff.js combines prettydiff barebones example files to 1:
			      https://github.com/prettydiff/prettydiff/blob/master/test/barebones/barebones.xhtml#L27
			      If you don't combine them then you'll have to enable "unsafe-eval" in Content Security Policy header in nginx.conf
			      as pagespeed_ngx will turn some of those .js files to inline code.
						Another option is to use https://github.com/prettier/prettier
		*/
		files = []string{
			addStaticPrefix(host, "minify/prettydiff.js"),
			addStaticPrefix(host, "minify/minify.js"),
		}
	}

	return files
}

func commafy(v interface{}) string {
	switch v := v.(type) {
	case int:
		return humanize.Comma(int64(v))
	case int64:
		return humanize.Comma(v)
	case float32, float64:
		return humanize.Commaf(v.(float64))
	default:
		log.Debug.Printf("unknown type %T\n", v)
		return ""
	}
}

var hmacSecret = func() string {
	return os.Getenv("hmac_secret")
}

// hmacKey generates an hmac key for our reverse image proxy
func hmacKey(u string) string {
	secret := hmacSecret()
	if secret == "" {
		log.Info.Println(`hmac secret for image proxy is blank. Please set the "hmac_secret" env variable`)
	}

	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write([]byte(u)); err != nil {
		log.Info.Println(err)
	}

	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

// join joins items in a slice
func join(sl ...string) string {
	var s []string
	for _, item := range sl {
		if item != "" {
			s = append(s, item)
		}
	}

	return strings.Join(s, ", ")
}

func jsonMarshal(v interface{}) template.JS {
	b, err := json.Marshal(v)
	if err != nil {
		log.Debug.Println("error:", err)
	}
	return template.JS(b)
}

var now = func() time.Time { return time.Now().UTC() }

func percent(v float64) string {
	return strconv.FormatFloat(v*100, 'f', 2, 64) + "%"
}

func safeHTML(value string) template.HTML {
	return template.HTML(value)
}

func sortWHOISNameServers(servers []whois.NameServer) []whois.NameServer {
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	return servers
}

func stripHTML(s string) string {
	p := strings.NewReader(s)
	doc, _ := goquery.NewDocumentFromReader(p)
	return doc.Text()
}

// source will show the source of an instant answer if data comes from a 3rd party
func source(answer instant.Data) string {
	var proxyFavIcon = func(u string) string {
		return fmt.Sprintf("/image/32x,s%v/%v", hmacKey(u), u)
	}

	var img string
	var f string

	switch answer.Type {
	case "breach":
		b := answer.Solution.(*breach.Response)
		switch b.Provider {
		case breach.HaveIBeenPwnedProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, breach.HaveIBeenPwnedProvider, proxyFavIcon("https://haveibeenpwned.com/favicon.ico"))
			f += fmt.Sprintf(`<br>%v <a href="https://haveibeenpwned.com/">%v</a>`, img, breach.HaveIBeenPwnedProvider)
		default:
			log.Debug.Printf("unknown breach provider %v\n", b.Provider)
		}
	case "congress":
		c := answer.Solution.(*congress.Response)
		switch c.Provider {
		case congress.ProPublicaProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, congress.ProPublicaProvider, proxyFavIcon("https://assets.propublica.org/prod/v3/images/favicon.ico"))
			f += fmt.Sprintf(`<br>%v <a href="https://www.propublica.org/">%v</a>`, img, congress.ProPublicaProvider)
		default:
			log.Debug.Printf("unknown congress provider %v\n", c.Provider)
		}
	case "discography":
		img = fmt.Sprintf(`<img width="12" height="12" alt="musicbrainz" src="%v"/>`, proxyFavIcon("https://musicbrainz.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://musicbrainz.org/">MusicBrainz</a>`, img)
	case "fedex":
		img = fmt.Sprintf(`<img width="12" height="12" alt="fedex" src="%v"/>`, proxyFavIcon("http://www.fedex.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.fedex.com">FedEx</a>`, img)
	case "currency":
		q := answer.Solution.(*instant.CurrencyResponse)
		switch q.ForexProvider {
		case currency.ECBProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, currency.ECBProvider, proxyFavIcon("http://www.ecb.europa.eu/favicon.ico"))
			f = fmt.Sprintf(`%v <a href="http://www.ecb.europa.eu/home/html/index.en.html">%v</a>`, img, currency.ECBProvider)
		default:
			log.Debug.Printf("unknown forex provider %v\n", q.ForexProvider)
		}
		switch q.CryptoProvider {
		case currency.CryptoCompareProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, currency.CryptoCompareProvider, proxyFavIcon("https://www.cryptocompare.com/media/20562/favicon.png?v=2"))
			f += fmt.Sprintf(`<br>%v <a href="https://www.cryptocompare.com/">%v</a>`, img, currency.CryptoCompareProvider)
		default:
			log.Debug.Printf("unknown cryptocurrency provider %v\n", q.CryptoProvider)
		}
	case "gdp", "population":
		var provider econ.Provider

		switch answer.Type {
		case "gdp":
			provider = answer.Solution.(*instant.GDPResponse).Provider
		case "population":
			provider = answer.Solution.(*instant.PopulationResponse).Provider
		}

		var makeSource = func(p econ.Provider) string {
			var f string

			switch p {
			case econ.TheWorldBankProvider:
				img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, econ.TheWorldBankProvider, proxyFavIcon("https://www.worldbank.org/content/dam/wbr-redesign/logos/wbg-favicon.png"))
				f += fmt.Sprintf(`%v <a href="https://www.worldbank.org/">%v</a>`, img, p)
			default:
				log.Debug.Printf("unknown population provider %v\n", p)
			}
			return f
		}

		f = makeSource(provider)
	case "stackoverflow":
		// TODO: I wasn't able to get both the User's display name and link to their profile or id.
		// Can select one or the other but not both in their filter.
		user := answer.Solution.(*instant.StackOverflowAnswer).Answer.User
		img = fmt.Sprintf(`<img width="12" height="12" alt="stackoverflow" src="%v"/>`, proxyFavIcon("https://cdn.sstatic.net/Sites/stackoverflow/img/favicon.ico"))
		f = fmt.Sprintf(`%v via %v <a href="https://stackoverflow.com/">Stack Overflow</a>`, user, img)
	case "status":
		img = fmt.Sprintf(`<img width="12" height="12" alt="isitup?" src="%v"/>`, proxyFavIcon("https://isitup.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://isitup.org/">Is It Up?</a>`, img)
	case "stock quote":
		q := answer.Solution.(*stock.Quote)
		switch q.Provider {
		case stock.IEXProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, q.Provider, proxyFavIcon("https://iextrading.com/favicon.ico"))
			f = fmt.Sprintf(`%v Data provided for free by <a href="https://iextrading.com/developer">%v</a>.`, img, q.Provider) // MUST say "Data provided for free by <a href="https://iextrading.com/developer">IEX</a>."
		default:
			log.Debug.Printf("unknown stock quote provider %v\n", q.Provider)
		}
	case "ups":
		img = fmt.Sprintf(`<img width="12" height="12" alt="ups" src="%v"/>`, proxyFavIcon("https://www.ups.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.ups.com">UPS</a>`, img)
	case "usps":
		img = fmt.Sprintf(`<img width="12" height="12" alt="usps" src="%v"/>`, proxyFavIcon("https://www.usps.com/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.usps.com">USPS</a>`, img)
	case "url shortener":
		s := answer.Solution.(*shortener.Response)
		switch s.Provider {
		case shortener.IsGdProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, shortener.IsGdProvider, proxyFavIcon("https://is.gd/isgd_favicon.ico"))
			f = fmt.Sprintf(`%v <a href="https://is.gd/">%v</a>`, img, shortener.IsGdProvider)
		default:
			log.Debug.Printf("unknown link shortening service %v\n", s.Provider)
		}
	case "local weather", "weather":
		w := answer.Solution.(*weather.Weather)
		switch w.Provider {
		case weather.OpenWeatherMapProvider:
			img = fmt.Sprintf(`<img width="12" height="12" alt="%v" src="%v"/>`, weather.OpenWeatherMapProvider, proxyFavIcon("http://openweathermap.org/favicon.ico"))
			f = fmt.Sprintf(`%v <a href="http://openweathermap.org">%v</a>`, img, weather.OpenWeatherMapProvider)
		default:
			log.Debug.Printf("unknown weather provider %v\n", w.Provider)
		}
	case "whois":
		img = fmt.Sprintf(`<img width="12" height="12" alt="jivedata" src="%v"/>`, proxyFavIcon("https://jivedata.com/static/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://jivedata.com">Jive Data</a>`, img)
	case "wikidata age", "wikidata clock", "wikidata birthday", "wikidata death", "wikidata height", "wikidata weight":
		img = fmt.Sprintf(`<img width="12" height="12" alt="wikipedia" src="%v"/>`, proxyFavIcon("https://en.wikipedia.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.wikipedia.org/">Wikipedia</a>`, img)
	case "wikiquote":
		img = fmt.Sprintf(`<img width="12" height="12" alt="wikiquote" src="%v"/>`, proxyFavIcon("https://en.wikiquote.org/favicon.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.wikiquote.org/">Wikiquote</a>`, img)
	case "wiktionary":
		img = fmt.Sprintf(`<img width="12" height="12" alt="wiktionary" src="%v"/>`, proxyFavIcon("https://www.wiktionary.org/static/favicon/piece.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.wiktionary.org/">Wiktionary</a>`, img)
	case "wikipedia":
		img = fmt.Sprintf(`<img width="12" height="12" alt="wikipedia" src="%v"/>`, proxyFavIcon("https://en.wikipedia.org/static/favicon/wikipedia.ico"))
		f = fmt.Sprintf(`%v <a href="https://www.wikipedia.org/">Wikipedia</a>`, img)
	default:
		f = "Jive Search"
	}

	return f
}

func subtract(x, y int) int {
	return x - y
}

func title(i interface{}) string {
	var s string

	switch i := i.(type) {
	case search.Filter:
		s = string(i)
	default:
		s = i.(string)
	}
	return strings.Title(s)
}

// Preserving words is a crude translation from the python answer:
// http://stackoverflow.com/questions/250357/truncate-a-string-without-ending-in-the-middle-of-a-word
func truncate(txt string, max int, preserve bool) string {
	if len(txt) <= max {
		return txt
	}

	if preserve {
		c := strings.Fields(txt[:max+1])
		return strings.Join(c[0:len(c)-1], " ") + " ..."
	}

	return txt[:max] + "..."
}

func weatherCode(c weather.Description) string {
	var icon string

	switch c {
	case weather.Clear:
		icon = "icon-sun"
	case weather.LightClouds:
		icon = "icon-cloud-sun"
	case weather.ScatteredClouds:
		icon = "icon-cloud"
	case weather.OvercastClouds:
		icon = "icon-cloud-inv"
	case weather.Extreme:
		icon = "icon-cloud-flash-inv"
	case weather.Rain:
		icon = "icon-rain"
	case weather.Snow:
		icon = "icon-snowflake-o"
	case weather.ThunderStorm:
		icon = "icon-cloud-flash"
	case weather.Windy:
		icon = "icon-windy"
	default:
		icon = "icon-sun"
	}

	return icon
}

type weatherDay struct {
	*weather.Instant
	DT    string
	codes map[weather.Description]int
}

// weatherDailyForecast combines multi-day weather forecasts to 1 daily forecast.
func weatherDailyForecast(forecasts []*weather.Instant, timezone string) []*weatherDay {
	tmp := map[string]*weatherDay{}
	dates := []time.Time{}
	days := []*weatherDay{}

	if timezone == "" { // this is just a hack until we can match timezones with zipcodes
		timezone = "America/Los_Angeles"
	}

	location, err := time.LoadLocation(timezone)
	if err != nil {
		log.Info.Println(err)
	}

	var fmtDate = func(d time.Time) string {
		return d.In(location).Format("Mon 02")
	}

	for _, f := range forecasts {
		fd := fmtDate(f.Date)

		if v, ok := tmp[fd]; ok {
			if f.High > v.Instant.High {
				v.Instant.High = f.High
			}
			if f.Low < v.Instant.Low {
				v.Instant.Low = f.Low
			}
			v.codes[f.Code]++
		} else {
			wd := &weatherDay{
				&weather.Instant{
					Date: f.Date,
					Low:  f.Low,
					High: f.High,
				}, fd, make(map[weather.Description]int),
			}
			wd.codes[f.Code]++
			tmp[fd] = wd
			dates = append(dates, f.Date)
		}
	}

	sort.Slice(dates, func(i, j int) bool { return dates[i].Before(dates[j]) })

	for _, d := range dates {
		f := fmtDate(d)
		// find the most frequently used icon for that day
		var most int
		for kk, v := range tmp[f].codes {
			if v > most {
				tmp[f].Code = kk
				most = v
			}
		}

		days = append(days, tmp[f])
	}

	return days
}

// wikiAmount displays a unit in meters, feet, etc depending on user's region
func wikiAmount(q wikipedia.Quantity, r language.Region) string {
	var f string

	amt, err := strconv.ParseFloat(q.Amount, 64)
	if err != nil {
		log.Debug.Println(err)
		return ""
	}

	switch r.String() {
	case "US", "LR", "MM": // only 3 countries that don't use metric system
		switch q.Unit.ID {
		case "Q11573", "Q174728", "Q218593":
			if q.Unit.ID == "Q11573" { // 1 meter = 39.3701 inches
				amt = amt * 39.3701
			} else if q.Unit.ID == "Q174728" { // 1 cm = 0.393701 inches
				amt = amt * .393701
			}

			if amt < 12 {
				f = fmt.Sprintf(`%f"`, amt)
			} else {
				f = fmt.Sprintf(`%d'%d"`, int(amt)/int(12), int(math.Mod(amt, 12)))
			}

		case "Q11570": // 1 kilogram = 2.20462 lbs
			amt = amt * 2.20462
			f = fmt.Sprintf("%d lbs", int(amt+.5))

		default:
			log.Debug.Printf("unknown unit %v\n", q.Unit.ID)
		}
	default:
		s := strconv.FormatFloat(amt, 'f', -1, 64)

		switch q.Unit.ID {
		case "Q11573":
			f = fmt.Sprintf("%v %v", s, "m")
		case "Q174728":
			f = fmt.Sprintf("%v %v", s, "cm")
		case "Q218593":
			amt = amt / .393701
			f = fmt.Sprintf("%v %v", int(amt+.5), "cm")
		case "Q11570":
			f = fmt.Sprintf("%v %v", s, "kg")
		default:
			log.Debug.Printf("unknown unit %v\n", q.Unit.ID)
		}
	}

	return f
}

// wikiCanonical returns the canonical form of a wikipedia title.
// if this breaks Wikidata dumps have "sitelinks"
func wikiCanonical(t string) string {
	return strings.Replace(t, " ", "_", -1)
}

func wikidataClockTime(t time.Time) string {
	return t.Format("3:04 PM")
}

func wikidataClockDate(t time.Time) string {
	return t.Format("Monday, January 2, 2006 (GMT-07)")
}

func wikiData(sol instant.Data, r language.Region) string {
	switch sol.Solution.(type) {
	case []wikipedia.Quantity: // height, weight, etc.
		i := sol.Solution.([]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case *[]wikipedia.Quantity: // cached version of height, weight, etc.
		i := *sol.Solution.(*[]wikipedia.Quantity)
		if len(i) == 0 {
			return ""
		}
		return wikiAmount(i[0], r)
	case *instant.Age:
		a := sol.Solution.(*instant.Age)

		// alive
		if a.Death == nil || reflect.DeepEqual(a.Death.Death, wikipedia.DateTime{}) {
			return fmt.Sprintf(`<em>Age:</em> %d Years<br><span style="color:#666;">%v</span>`,
				wikiYears(a.Birthday.Birthday, now()), wikiDateTime(a.Birthday.Birthday))
		}

		// dead
		return fmt.Sprintf(`<em>Age at Death:</em> %d Years<br><span style="color:#666;">%v - %v</span>`,
			wikiYears(a.Birthday.Birthday, a.Death.Death), wikiDateTime(a.Birthday.Birthday), wikiDateTime(a.Death.Death))
	case *instant.Birthday:
		b := sol.Solution.(*instant.Birthday)
		return wikiDateTime(b.Birthday)
	case *instant.Death:
		d := sol.Solution.(*instant.Death)
		return wikiDateTime(d.Death)
	default:
		log.Debug.Printf("unknown instant solution type %T\n", sol.Solution)
		return ""
	}
}

// wikiDateTime formats a date with optional time.
// We assume Gregorian calendar below. (Julian calendar TODO).
// Note: Wikidata only uses Gregorian and Julian calendars.
func wikiDateTime(dt wikipedia.DateTime) string {
	// we loop through the formats until one is found
	// starting with most specific and ending with most general order
	for j, f := range []string{time.RFC3339Nano, "2006"} {
		var ff string

		switch j {
		case 1:
			dt.Value = dt.Value[:4]
			ff = f
		default:
			ff = "January 2, 2006"
		}

		t, err := time.Parse(f, dt.Value)
		if err != nil {
			log.Debug.Println(err)
			continue
		}

		return t.Format(ff)
	}

	return ""
}

func wikipediaItem(sol instant.Data) []*wikipedia.Item {
	if sol.Solution == nil {
		return []*wikipedia.Item{}
	}
	return sol.Solution.([]*wikipedia.Item)
}

// wikiJoin joins a slice of Wikidata items
func wikiJoin(items []wikipedia.Wikidata, preferred []language.Tag) string {
	sl := []string{}
	for _, item := range items {
		sl = append(sl, wikiLabel(item.Labels, preferred))
	}

	return strings.Join(sl, ", ")
}

// wikiLabel extracts the closest label for a Wikipedia Item using a language matcher
func wikiLabel(labels map[string]wikipedia.Text, preferred []language.Tag) string {
	// create a matcher based on the available labels
	langs := []language.Tag{}

	for k := range labels {
		t, err := language.Parse(k)
		if err != nil { // sr-el doesn't parse
			continue
		}

		langs = append(langs, t)
	}

	m := language.NewMatcher(langs)
	lang, _, _ := m.Match(preferred...)

	label := labels[lang.String()]
	return label.Text
}

// wikiYears calculates the number of years (rounded down) betwee two dates.
// e.g. a person's age
func wikiYears(start, end interface{}) int {
	var parseDateTime = func(d interface{}) time.Time {
		switch d := d.(type) {
		case wikipedia.DateTime:
			for j, f := range []string{time.RFC3339Nano, "2006"} {
				if j == 1 {
					d.Value = d.Value[:4]
				}
				t, err := time.Parse(f, d.Value)
				if err != nil {
					log.Debug.Println(err)
					continue
				}
				return t
			}

		case time.Time:
			return d
		default:
			log.Debug.Printf("unknown type %T\n", d)
		}
		return time.Time{}
	}

	s := parseDateTime(start)
	e := parseDateTime(end)

	years := e.Year() - s.Year()
	if e.YearDay() < s.YearDay() {
		years--
	}

	return years
}
