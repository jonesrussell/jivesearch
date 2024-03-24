package currency

import (
	"time"

	"github.com/openprovider/ecbrates"
)

// ECBProvider is a currency provider
var ECBProvider provider = "European Central Bank"

// ECB holds settings for ECB via openprovider
type ECB struct{}

// Fetch retrieves currency quotes from the ECB via openprovider
func (e *ECB) Fetch() (*Response, error) {
	//rates, err := ecbrates.LoadAll() // full history...takes longer (make method configurable)
	rates, err := ecbrates.Load() // 90 days...sufficient?
	if err != nil {
		return nil, err
	}

	resp := New()
	resp.ForexProvider = ECBProvider

	// just grab all the rates...why not?
	for _, r := range rates {
		d, err := time.Parse("2006-01-02", r.Date)
		if err != nil {
			return nil, err
		}

		for c := range r.Rate {
			var currency Currency

			if ok, cc := Valid(string(c)); ok {
				currency = cc
			}

			rate := &Rate{
				DateTime: d,
			}

			rate.Rate, err = r.Convert(1, c, ecbrates.Currency(resp.Base.Short))
			if err != nil {
				return nil, err
			}

			resp.History[currency.Short] = append(resp.History[currency.Short], rate)
		}
	}

	return resp, err
}
