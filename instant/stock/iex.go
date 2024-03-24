package stock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// IEX retrieves information from the IEX API
type IEX struct {
	HTTPClient *http.Client
}

// IEXProvider is a stock quote provider
var IEXProvider provider = "IEX"

type iexResponse struct {
	Quote    *Quote
	RawQuote struct {
		Symbol           string      `json:"symbol"`
		CompanyName      string      `json:"companyName"`
		PrimaryExchange  string      `json:"primaryExchange"`
		Sector           string      `json:"sector"`
		CalculationPrice string      `json:"calculationPrice"`
		Open             float64     `json:"open"`
		OpenTime         int64       `json:"openTime"`
		Close            float64     `json:"close"`
		CloseTime        int64       `json:"closeTime"`
		High             float64     `json:"high"`
		Low              float64     `json:"low"`
		LatestPrice      float64     `json:"latestPrice"`
		LatestSource     string      `json:"latestSource"`
		LatestTime       string      `json:"latestTime"`
		LatestUpdate     int64       `json:"latestUpdate"`
		LatestVolume     int         `json:"latestVolume"`
		IexRealtimePrice interface{} `json:"iexRealtimePrice"`
		IexRealtimeSize  interface{} `json:"iexRealtimeSize"`
		IexLastUpdated   interface{} `json:"iexLastUpdated"`
		DelayedPrice     float64     `json:"delayedPrice"`
		DelayedPriceTime int64       `json:"delayedPriceTime"`
		PreviousClose    float64     `json:"previousClose"`
		Change           float64     `json:"change"`
		ChangePercent    float64     `json:"changePercent"`
		IexMarketPercent interface{} `json:"iexMarketPercent"`
		IexVolume        interface{} `json:"iexVolume"`
		AvgTotalVolume   int         `json:"avgTotalVolume"`
		IexBidPrice      interface{} `json:"iexBidPrice"`
		IexBidSize       interface{} `json:"iexBidSize"`
		IexAskPrice      interface{} `json:"iexAskPrice"`
		IexAskSize       interface{} `json:"iexAskSize"`
		MarketCap        int64       `json:"marketCap"`
		PeRatio          float64     `json:"peRatio"`
		Week52High       float64     `json:"week52High"`
		Week52Low        float64     `json:"week52Low"`
		YtdChange        float64     `json:"ytdChange"`
	} `json:"quote"`
	Chart []struct {
		Date             string  `json:"date"`
		Open             float64 `json:"open"`
		High             float64 `json:"high"`
		Low              float64 `json:"low"`
		Close            float64 `json:"close"`
		Volume           int     `json:"volume"`
		UnadjustedVolume int     `json:"unadjustedVolume"`
		Change           float64 `json:"change"`
		ChangePercent    float64 `json:"changePercent"`
		Vwap             float64 `json:"vwap"`
		Label            string  `json:"label"`
		ChangeOverTime   float64 `json:"changeOverTime"`
	} `json:"chart"`
}

// UnmarshalJSON sets the Response fields
func (r *iexResponse) UnmarshalJSON(b []byte) error {
	type alias iexResponse
	raw := &alias{}

	err := json.Unmarshal(b, &raw)
	if err != nil {
		return err
	}

	location, err := time.LoadLocation("America/New_York")
	if err != nil {
		return err
	}

	r.Quote = &Quote{
		Provider: IEXProvider,
		Ticker:   raw.RawQuote.Symbol,
		Name:     raw.RawQuote.CompanyName,
		Last: Last{
			Price:         raw.RawQuote.LatestPrice,
			Time:          time.Unix(raw.RawQuote.LatestUpdate/1000, 0).In(location),
			Change:        raw.RawQuote.Change,
			ChangePercent: raw.RawQuote.ChangePercent,
		},
	}

	r.Quote, err = r.Quote.exchange(raw.RawQuote.PrimaryExchange)
	if err != nil {
		return err
	}

	for _, v := range raw.Chart {
		dt, err := time.Parse("2006-01-02", v.Date)
		if err != nil {
			return err
		}

		q := EOD{
			Date:   dt,
			Open:   v.Open,
			Close:  v.Close,
			High:   v.High,
			Low:    v.Low,
			Volume: v.Volume,
		}

		r.Quote.History = append(r.Quote.History, q)
	}

	return err
}

// Fetch retrieves from the IEX api
func (i *IEX) Fetch(ticker string) (*Quote, error) {
	iex := iexResponse{}

	u := fmt.Sprintf("https://api.iextrading.com/1.0/stock/%s/batch?types=quote,chart&range=5y", ticker)

	resp, err := i.HTTPClient.Get(u)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil, ErrInvalidTicker
	}

	err = json.NewDecoder(resp.Body).Decode(&iex)

	return iex.Quote, err
}
