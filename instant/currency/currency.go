// Package currency fetches foreign exchange quotes
package currency

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// cases
// 1: "usd to eur"
// 2: "12 usd to jpy"
// 3: "convert 1 usd to jpy"

// FXFetcher retrieves fx quotes
type FXFetcher interface {
	Fetch() (*Response, error)
}

// CryptoFetcher retrieves cryptocurrency quotes
type CryptoFetcher interface {
	Fetch() (*Response, error)
}

type provider string

// ErrInvalidCurrency indicates the currency was invalid
var ErrInvalidCurrency = fmt.Errorf("invalid currency")

// Rate is an currency quote
type Rate struct {
	DateTime time.Time
	Rate     float64
}

// Response is a currency response
type Response struct {
	Base           Currency // the base currency
	History        map[string][]*Rate
	ForexProvider  provider
	CryptoProvider provider
}

// New returns a new Response with the base currency set to USD
func New() *Response {
	return &Response{
		Base:    USD,
		History: make(map[string][]*Rate),
	}
}

// Sort sorts the historical rates by date in Asc order
func (r *Response) Sort() *Response {
	for k, v := range r.History {
		sort.Slice(v, func(i, j int) bool {
			return r.History[k][i].DateTime.Before(r.History[k][j].DateTime)
		})
	}

	return r
}

// Currency is an FX currency
type Currency struct {
	Short string
	Long  string
}

// fx currencies
var (
	// USD is a Currency
	USD = Currency{"USD", "US Dollar"}
	// AUD is a Currency
	AUD = Currency{"AUD", "Australian Dollar"}
	// BGN is a Currency
	BGN = Currency{"BGN", "Bulgarian Lev"}
	// BRL is a Currency
	BRL = Currency{"BRL", "Brazilian Real"}
	// CAD is a Currency
	CAD = Currency{"CAD", "Canadian Dollar"}
	// CHF is a Currency
	CHF = Currency{"CHF", "Swiss Franc"}
	// CNY is a Currency
	CNY = Currency{"CNY", "Chinese Yuan"}
	// CZK is a Currency
	CZK = Currency{"CZK", "Czech Republic Koruna"}
	// DKK is a Currency
	DKK = Currency{"DKK", "Danish Krone"}
	// EUR is a Currency
	EUR = Currency{"EUR", "Euro"}
	// GBP is a Currency
	GBP = Currency{"GBP", "British Pound Sterling"}
	// HKD is a Currency
	HKD = Currency{"HKD", "Hong Kong Dollar"}
	// HRK is a Currency
	HRK = Currency{"HRK", "Croatian Kuna"}
	// HUF is a Currency
	HUF = Currency{"HUF", "Hungarian Forint"}
	// IDR is a Currency
	IDR = Currency{"IDR", "Indonesian Rupiah"}
	// ILS is a Currency
	ILS = Currency{"ILS", "Israeli New Sheqel"}
	// INR is a Currency
	INR = Currency{"INR", "Indian Rupee"}
	// ISK is a Currency
	ISK = Currency{"ISK", "Iceland Krona"}
	// JPY is a Currency
	JPY = Currency{"JPY", "Japanese Yen"}
	// KRW is a Currency
	KRW = Currency{"KRW", "South Korean Won"}
	// LTL is a Currency
	LTL = Currency{"LTL", "Lithuanian Litas"}
	// MXN is a Currency
	MXN = Currency{"MXN", "Mexican Peso"}
	// MYR is a Currency
	MYR = Currency{"MYR", "Malaysian Ringgit"}
	// NOK is a Currency
	NOK = Currency{"NOK", "Norwegian Krone"}
	// NZD is a Currency
	NZD = Currency{"NZD", "New Zealand Dollar"}
	// PHP is a Currency
	PHP = Currency{"PHP", "Philippine Peso"}
	// PLN is a Currency
	PLN = Currency{"PLN", "Polish Zloty"}
	// RON is a Currency
	RON = Currency{"RON", "Romanian Leu"}
	// RUB is a Currency
	RUB = Currency{"RUB", "Russian Ruble"}
	// SEK is a Currency
	SEK = Currency{"SEK", "Swedish Krona"}
	// SGD is a Currency
	SGD = Currency{"SGD", "Singapore Dollar"}
	// THB is a Currency
	THB = Currency{"THB", "Thai Baht"}
	// TRY is a Currency
	TRY = Currency{"TRY", "Turkish Lira"}
	// ZAR is a Currency
	ZAR = Currency{"ZAR", "South African Rand"}
)

// crypto currencies
var (
	BTC  = Currency{"BTC", "Bitcoin"}
	DOGE = Currency{"DOGE", "Dogecoin"}
	ETH  = Currency{"ETH", "Ethereum"}
	LTC  = Currency{"LTC", "Litecoin"}
	XMR  = Currency{"XMR", "Monero"}
	XRP  = Currency{"XRP", "Ripple"}
)

// ForexCurrencies are valid forex currencies
var ForexCurrencies = []Currency{
	AUD,
	BGN,
	BRL,
	CAD,
	CHF,
	CNY,
	CZK,
	DKK,
	EUR,
	GBP,
	HKD,
	HRK,
	HUF,
	IDR,
	ILS,
	INR,
	ISK,
	JPY,
	KRW,
	LTL,
	MXN,
	MYR,
	NOK,
	NZD,
	PHP,
	PLN,
	RON,
	RUB,
	SEK,
	SGD,
	THB,
	TRY,
	USD,
	ZAR,
}

// CryptoCurrencies are valid crypto currencies
var CryptoCurrencies = []Currency{
	BTC,
	DOGE,
	ETH,
	LTC,
	XMR,
	XRP,
}

// Valid checks if a given currency is supported
func Valid(c string) (bool, Currency) {
	ac := append(ForexCurrencies, CryptoCurrencies...)

	for _, cu := range ac {
		if strings.EqualFold(c, cu.Short) {
			return true, cu
		}
	}

	return false, Currency{}
}
