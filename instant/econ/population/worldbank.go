package population

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/jonesrussell/jivesearch/instant/econ"
)

// WorldBank holds settings for The World Bank API
type WorldBank struct {
	HTTPClient *http.Client
}

// Fetch retrieves population data from The World Bank
func (w *WorldBank) Fetch(country string, from, to time.Time) (*Response, error) {
	u, err := w.buildURL(country, from, to)
	if err != nil {
		return nil, err
	}

	resp, err := w.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	wr := &econ.WorldBankResponse{}

	if err = xml.Unmarshal(bdy, &wr); err != nil {
		return nil, err
	}

	r := &Response{
		Provider: econ.TheWorldBankProvider,
	}

	for _, pp := range wr.Data {
		i := Instant{
			Date:  time.Date(pp.Date.Date, 12, 31, 0, 0, 0, 0, time.UTC),
			Value: pp.Value.Value,
		}

		if i.Value == 0 {
			continue
		}

		r.History = append(r.History, i)
	}

	return r, err
}

func (w *WorldBank) buildURL(country string, from, to time.Time) (*url.URL, error) {
	// http://api.worldbank.org/v2/countries/it/indicators/SP.POP.TOTL?date=1960:2018
	u, err := url.Parse(fmt.Sprintf("http://api.worldbank.org/v2/countries/%v/indicators/SP.POP.TOTL", country))
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("date", fmt.Sprintf("%d:%d", from.Year(), to.Year()))
	u.RawQuery = q.Encode()

	return u, err
}
