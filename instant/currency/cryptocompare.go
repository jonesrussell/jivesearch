package currency

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// CryptoCompareProvider is a currency provider
var CryptoCompareProvider provider = "CryptoCompare"

// CryptoCompare holds settings for the CryptoCompare api
// CryptoCompare allows 8K historical requests/hr (133/minute).
type CryptoCompare struct {
	*http.Client
	UserAgent string
}

// CryptoCompareResponse is the raw CryptoCompare response
type CryptoCompareResponse struct {
	Response   string `json:"Response"`
	Type       int    `json:"Type"`
	Aggregated bool   `json:"Aggregated"`
	Data       []struct {
		Time       int     `json:"time"`
		High       float64 `json:"high"`
		Low        float64 `json:"low"`
		Open       float64 `json:"open"`
		Volumefrom float64 `json:"volumefrom"`
		Volumeto   float64 `json:"volumeto"`
		Close      float64 `json:"close"`
	} `json:"Data"`
	TimeTo            int  `json:"TimeTo"`
	TimeFrom          int  `json:"TimeFrom"`
	FirstValueInArray bool `json:"FirstValueInArray"`
	ConversionType    struct {
		Type             string `json:"type"`
		ConversionSymbol string `json:"conversionSymbol"`
	} `json:"ConversionType"`
}

// Fetch retrieves a cryptocurrency quotes from CryptoCompare
func (c *CryptoCompare) Fetch() (*Response, error) {
	var err error

	type tmp struct {
		Currency
		hist []*Rate
	}

	ch := make(chan tmp)
	errCh := make(chan error)

	// Have to retrieve each cryptocurrency history separately
	// since CryptoCompare does not have an endpoint to fetch all.
	// Note: We only have to fetch the "from" and "to" ("LTC to BTC") but might as
	// well get all of cryptos since they are fetched concurrently.
	for _, cur := range CryptoCurrencies {
		go func(cur Currency) {
			t := tmp{
				Currency: cur,
			}

			cr, err := c.fetch(cur, USD)
			if err != nil {
				errCh <- err
			}

			for _, d := range cr.Data {
				i, err := strconv.ParseInt(strconv.Itoa(d.Time), 10, 64)
				if err != nil {
					errCh <- err
				}
				tm := time.Unix(i, 0).In(time.UTC)

				rate := &Rate{
					DateTime: tm,
					Rate:     d.Close,
				}

				t.hist = append(t.hist, rate)
			}

			ch <- t
		}(cur)
	}

	resp := New()
	resp.CryptoProvider = CryptoCompareProvider

	for range CryptoCurrencies {
		select {
		case t := <-ch:
			resp.History[t.Currency.Short] = t.hist
		case err := <-errCh:
			return nil, err
		}
	}

	return resp, err
}

func (c *CryptoCompare) buildURL(from, to Currency) (*url.URL, error) {
	//https://min-api.cryptocompare.com/data/histoday?fsym=BTC&tsym=USD&limit=60&aggregate=3&e=CCCAGG
	u, err := url.Parse("https://min-api.cryptocompare.com/data/histoday")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("fsym", from.Short)
	q.Add("tsym", to.Short)
	q.Add("limit", "90")
	q.Add("extraParams", c.UserAgent)
	u.RawQuery = q.Encode()

	return u, err
}

func (c *CryptoCompare) fetch(from, to Currency) (*CryptoCompareResponse, error) {
	u, err := c.buildURL(from, to)
	if err != nil {
		return nil, err
	}

	resp, err := c.Client.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	cr := &CryptoCompareResponse{}
	err = json.NewDecoder(resp.Body).Decode(&cr)

	return cr, err
}
