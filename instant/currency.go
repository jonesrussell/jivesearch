package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	curr "github.com/jonesrussell/jivesearch/instant/currency"
	"golang.org/x/text/language"
)

// CurrencyType is an answer Type
const CurrencyType Type = "currency"

// Currency is an instant answer
type Currency struct {
	FXFetcher     curr.FXFetcher
	CryptoFetcher curr.CryptoFetcher
	Answer
}

// CurrencyResponse is an instant answer response
type CurrencyResponse struct {
	*curr.Response
	Notional         float64
	From             curr.Currency
	To               curr.Currency
	ForexCurrencies  []curr.Currency
	CryptoCurrencies []curr.Currency
	Currencies       []curr.Currency
}

// ErrInvalidCurrency indicates the currency was invalid
var ErrInvalidCurrency = fmt.Errorf("invalid currency")

func (c *Currency) setQuery(r *http.Request, qv string) Answerer {
	c.Answer.setQuery(r, qv)
	return c
}

func (c *Currency) setUserAgent(r *http.Request) Answerer {
	return c
}

func (c *Currency) setLanguage(lang language.Tag) Answerer {
	c.language = lang
	return c
}

func (c *Currency) setType() Answerer {
	c.Type = CurrencyType
	return c
}

func (c *Currency) setRegex() Answerer {
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<notional>\d+) (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^convert (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<notional>\d+) (?P<from>.*)$`))

	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*) to (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*) (?P<to>.*)$`))
	c.regex = append(c.regex, regexp.MustCompile(`^(?P<from>.*)$`))

	return c
}

func (c *Currency) solve(r *http.Request) Answerer {
	resp := &CurrencyResponse{
		Response:         curr.New(),
		Notional:         1,
		Currencies:       append(curr.ForexCurrencies, curr.CryptoCurrencies...),
		ForexCurrencies:  curr.ForexCurrencies,
		CryptoCurrencies: curr.CryptoCurrencies,
	}

	if n, ok := c.remainderM["notional"]; ok {
		nn, err := strconv.ParseFloat(n, 64)
		if err != nil {
			c.Err = err
			return c
		}
		resp.Notional = nn
	}

	var ok bool

	from := c.remainderM["from"]
	ok, resp.From = curr.Valid(from)
	if !ok {
		c.Err = ErrInvalidCurrency
		return c
	}

	to := c.remainderM["to"]
	ok, resp.To = curr.Valid(to)
	if !ok {
		if resp.From == curr.PHP { // chances are they are looking for a programming answer for PHP
			c.Err = ErrInvalidCurrency
			return c
		}
		resp.To = curr.USD // assume USD for second if not specified "125 BTC"
	}

	cch := make(chan *curr.Response)
	fch := make(chan *curr.Response)
	ech := make(chan error)

	go func(ch chan *curr.Response) {
		crytopResp, err := c.CryptoFetcher.Fetch()
		if err != nil {
			ech <- err
		}
		cch <- crytopResp
	}(cch)

	go func(ch chan *curr.Response) {
		forexResp, err := c.FXFetcher.Fetch()
		if err != nil {
			ech <- err
		}
		fch <- forexResp
	}(fch)

	cresp := &curr.Response{}

	for i := 0; i <= 1; i++ {
		select {
		case res := <-cch:
			cresp.CryptoProvider = res.CryptoProvider
			cresp.History = res.History
		case res := <-fch:
			resp.ForexProvider = res.ForexProvider
			resp.History = res.History
		case err := <-ech:
			c.Err = err
			return c
		}
	}

	resp.CryptoProvider = cresp.CryptoProvider

	// Add crypto to the history
	for k, v := range cresp.History {
		resp.History[k] = v
	}

	resp.Sort()

	c.Data.Solution = resp
	return c
}

func (c *Currency) tests() []test {
	history := map[string][]*curr.Rate{
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
	}

	tests := []test{
		{
			query: "convert JPY to USD",
			expected: []Data{
				{
					Type:      CurrencyType,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &curr.Response{
							Base:           curr.USD,
							History:        history,
							CryptoProvider: curr.CryptoCompareProvider,
							ForexProvider:  curr.ECBProvider,
						},
						Notional:         1,
						From:             curr.JPY,
						To:               curr.USD,
						Currencies:       append(curr.ForexCurrencies, curr.CryptoCurrencies...),
						ForexCurrencies:  curr.ForexCurrencies,
						CryptoCurrencies: curr.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "125 EUR to JPY",
			expected: []Data{
				{
					Type:      CurrencyType,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &curr.Response{
							Base:           curr.USD,
							History:        history,
							CryptoProvider: curr.CryptoCompareProvider,
							ForexProvider:  curr.ECBProvider,
						},
						Notional:         125,
						From:             curr.EUR,
						To:               curr.JPY,
						Currencies:       append(curr.ForexCurrencies, curr.CryptoCurrencies...),
						ForexCurrencies:  curr.ForexCurrencies,
						CryptoCurrencies: curr.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "BTC",
			expected: []Data{
				{
					Type:      CurrencyType,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &curr.Response{
							Base:           curr.USD,
							History:        history,
							CryptoProvider: curr.CryptoCompareProvider,
							ForexProvider:  curr.ECBProvider,
						},
						Notional:         1,
						From:             curr.BTC,
						To:               curr.USD,
						Currencies:       append(curr.ForexCurrencies, curr.CryptoCurrencies...),
						ForexCurrencies:  curr.ForexCurrencies,
						CryptoCurrencies: curr.CryptoCurrencies,
					},
				},
			},
		},
		{
			query: "125 BTC",
			expected: []Data{
				{
					Type:      CurrencyType,
					Triggered: true,
					Solution: &CurrencyResponse{
						Response: &curr.Response{
							Base:           curr.USD,
							History:        history,
							CryptoProvider: curr.CryptoCompareProvider,
							ForexProvider:  curr.ECBProvider,
						},
						Notional:         125,
						From:             curr.BTC,
						To:               curr.USD,
						Currencies:       append(curr.ForexCurrencies, curr.CryptoCurrencies...),
						ForexCurrencies:  curr.ForexCurrencies,
						CryptoCurrencies: curr.CryptoCurrencies,
					},
				},
			},
		},
	}

	return tests
}
