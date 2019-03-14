// Package instant provides instant answers
package instant

import (
	"errors"
	"math/rand"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/breach"
	"github.com/jivesearch/jivesearch/instant/congress"
	"github.com/jivesearch/jivesearch/instant/nutrition"
	"github.com/jivesearch/jivesearch/instant/status"
	"github.com/jivesearch/jivesearch/instant/timezone"
	"github.com/jivesearch/jivesearch/instant/whois"

	ggdp "github.com/jivesearch/jivesearch/instant/econ/gdp"

	disc "github.com/jivesearch/jivesearch/instant/discography"
	pop "github.com/jivesearch/jivesearch/instant/econ/population"
	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/parcel"
	"github.com/jivesearch/jivesearch/instant/shortener"
	so "github.com/jivesearch/jivesearch/instant/stackoverflow"
	"github.com/jivesearch/jivesearch/instant/stock"
	"github.com/jivesearch/jivesearch/instant/weather"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

// Instant holds config information for the instant answers
type Instant struct {
	QueryVar           string
	BreachFetcher      breach.Fetcher
	CongressFetcher    congress.Fetcher
	DiscographyFetcher disc.Fetcher
	FedExFetcher       parcel.Fetcher
	Currency
	GDPFetcher           ggdp.Fetcher
	LinkShortener        shortener.Service
	LocationFetcher      location.Fetcher
	NutritionFetcher     nutrition.Fetcher
	PopulationFetcher    pop.Fetcher
	StackOverflowFetcher so.Fetcher
	StatusFetcher        status.Fetcher
	StockQuoteFetcher    stock.Fetcher
	TimeZoneFetcher      timezone.Fetcher
	UPSFetcher           parcel.Fetcher
	USPSFetcher          parcel.Fetcher
	WeatherFetcher       weather.Fetcher
	WHOISFetcher         whois.Fetcher
	WikipediaFetcher     wikipedia.Fetcher
}

// Answerer outlines methods for an instant answer
type Answerer interface {
	setQuery(r *http.Request, qv string) Answerer
	setUserAgent(r *http.Request) Answerer
	setLanguage(lang language.Tag) Answerer
	setType() Answerer
	setRegex() Answerer
	trigger() bool
	solve(r *http.Request) Answerer
	solution() Data
	tests() []test
}

// Answer holds an instant answer when triggered
type Answer struct {
	query       string
	userAgent   string
	language    language.Tag
	regex       []*regexp.Regexp
	triggerWord string
	remainder   string
	remainderM  map[string]string
	Data
}

// Type is the answer type
type Type string

// For mocking....MUST use time.Date(2016, 6, 5, 3, 2, 0, 0, time.UTC) in tests
var now = func() time.Time { return time.Now().UTC() }

// Data holds the returned data of an answer
type Data struct {
	Type      `json:"type,omitempty"`
	Triggered bool        `json:"triggered"`
	Solution  interface{} `json:"answer,omitempty"`
	Err       error       `json:"-"`
}

// Triggerer detects if the answer has been triggered
type Triggerer interface {
	Trigger()
}

// Trigger will trigger an instant answer
func (i *Instant) Trigger(ia Answerer, r *http.Request, lang language.Tag) bool {
	ia.setUserAgent(r).setQuery(r, i.QueryVar).setLanguage(lang).setRegex()
	return ia.trigger()
}

// Solve solves an instant answer
func (i *Instant) Solve(ia Answerer, r *http.Request) Data {
	ia.setType().solve(r)
	return ia.solution()
}

// setQuery sets the query field
func (a *Answer) setQuery(r *http.Request, qv string) {
	q := strings.ToLower(strings.TrimSpace(r.FormValue(qv)))
	q = strings.Trim(q, "?")
	a.query = strings.Join(strings.Fields(q), " ") // Replace multiple whitespace w/ single whitespace
}

func getIPAddress(r *http.Request) net.IP {
	maxCidrBlocks := []string{
		"127.0.0.1/8", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16",
		"169.254.0.0/16", "::1/128", "fc00::/7", "fe80::/10",
	}

	cidrs := make([]*net.IPNet, len(maxCidrBlocks))
	for i, maxCidrBlock := range maxCidrBlocks {
		_, cidr, _ := net.ParseCIDR(maxCidrBlock)
		cidrs[i] = cidr
	}

	isPrivateAddress := func(address string) (bool, error) {
		ipAddress := net.ParseIP(address)
		if ipAddress == nil {
			return false, errors.New("is private address")
		}

		for i := range cidrs {
			if cidrs[i].Contains(ipAddress) {
				return true, nil
			}
		}

		return false, nil
	}

	// Check X-Forward-For and X-Real-IP first
	var ip string
	for _, h := range []string{"X-Forwarded-For", "X-Real-IP"} {
		for _, address := range strings.Split(r.Header.Get(h), ",") {
			ip = strings.TrimSpace(address)
			isPrivate, err := isPrivateAddress(ip)
			if !isPrivate && err == nil {
				return net.ParseIP(ip)
			}
		}
	}

	ip = r.RemoteAddr
	if strings.ContainsRune(r.RemoteAddr, ':') {
		ip, _, _ = net.SplitHostPort(r.RemoteAddr)
	}

	return net.ParseIP(ip)
}

// trigger executes the regex for an instant answer
func (a *Answer) trigger() bool {
	a.remainderM = map[string]string{}

	for _, re := range a.regex {
		match := re.FindStringSubmatch(a.query)
		if len(match) == 0 {
			continue
		}

		for i, name := range re.SubexpNames() {
			if i == 0 || name == "" {
				continue
			}
			a.Triggered = true

			switch name {
			case "trigger":
				a.triggerWord = match[i]
			case "remainder":
				a.remainder = match[i]
			default:
				a.remainderM[name] = match[i]
			}
		}
		break
	}
	return a.Triggered
}

func (a *Answer) solution() Data {
	return a.Data
}

type test struct {
	query     string
	userAgent string
	ip        net.IP
	expected  []Data
}

func init() {
	// for coin, random & probably others down the road
	rand.Seed(time.Now().UTC().UnixNano())
}
