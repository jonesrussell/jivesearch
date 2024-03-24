package stock

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestIEXFetch(t *testing.T) {
	type args struct {
		ticker string
	}

	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		t.Fatal(err)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		u    string
		resp string
		want *Quote
	}{
		{
			name: "AAPL Quote",
			args: args{"AAPL"},
			u:    `https://api.iextrading.com/1.0/stock/AAPL/batch?types=quote,chart&range=5y`,
			resp: `{"quote":{"symbol":"AAPL","companyName":"Apple Inc.","primaryExchange":"Nasdaq Global Select","sector":"Technology","calculationPrice":"tops","open":168.07,"openTime":1522071000233,"close":164.94,"closeTime":1521835200671,"high":171.59,"low":166.44,"latestPrice":171.42,"latestSource":"IEX real time price","latestTime":"2:52:35 PM","latestUpdate":1522090355062,"latestVolume":25265896,"iexRealtimePrice":171.42,"iexRealtimeSize":100,"iexLastUpdated":1522090355062,"delayedPrice":171.533,"delayedPriceTime":1522089470450,"previousClose":164.94,"change":6.48,"changePercent":0.03929,"iexMarketPercent":0.03243,"iexVolume":819373,"avgTotalVolume":36742833,"iexBidPrice":171.42,"iexBidSize":100,"iexAskPrice":171.45,"iexAskSize":100,"marketCap":869787308460,"peRatio":18.63,"week52High":183.5,"week52Low":138.62,"ytdChange":-0.038589885200847475},"chart":[{"date":"2013-03-26","open":60.5276,"high":60.5797,"low":59.8891,"close":59.9679,"volume":73428208,"unadjustedVolume":10489744,"change":-0.317828,"changePercent":-0.527,"vwap":60.1238,"label":"Mar 26, 13","changeOverTime":0},{"date":"2013-03-27","open":59.3599,"high":59.4041,"low":58.6147,"close":58.7903,"volume":81854409,"unadjustedVolume":11693487,"change":-1.1777,"changePercent":-1.964,"vwap":58.9435,"label":"Mar 27, 13","changeOverTime":-0.019637172553983017}]}`,
			want: &Quote{
				Ticker:   "AAPL",
				Name:     "Apple Inc.",
				Exchange: NASDAQ,
				Last: Last{
					Price:         171.42,
					Time:          time.Unix(1522090355062/1000, 0).In(est),
					Change:        6.48,
					ChangePercent: 0.03929,
				},
				History: []EOD{
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
				},
				Provider: IEXProvider,
			},
		},
		{
			name: "$GE",
			args: args{"GE"},
			u:    `https://api.iextrading.com/1.0/stock/GE/batch?types=quote,chart&range=5y`,
			resp: `{"quote":{"symbol":"GE","companyName":"General Electric Company","primaryExchange":"New York Stock Exchange","sector":"Industrials","calculationPrice":"close","open":13.2,"openTime":1522071021640,"close":12.89,"closeTime":1522094417778,"high":13.239,"low":12.73,"latestPrice":12.89,"latestSource":"Close","latestTime":"March 26, 2018","latestUpdate":1522094417778,"latestVolume":101836295,"iexRealtimePrice":12.895,"iexRealtimeSize":100,"iexLastUpdated":1522094398034,"delayedPrice":12.93,"delayedPriceTime":1522097957811,"previousClose":13.07,"change":-0.18,"changePercent":-0.01377,"iexMarketPercent":0.02386,"iexVolume":2429814,"avgTotalVolume":79805405,"iexBidPrice":0,"iexBidSize":0,"iexAskPrice":0,"iexAskSize":0,"marketCap":111930473998,"peRatio":12.28,"week52High":30.54,"week52Low":12.89,"ytdChange":-0.2670644444942913},"chart":[{"date":"2013-03-26","open":19.7391,"high":19.7645,"low":19.5185,"close":19.6118,"volume":32353323,"unadjustedVolume":32353323,"change":-0.101793,"changePercent":-0.516,"vwap":19.6314,"label":"Mar 26, 13","changeOverTime":0},{"date":"2013-03-27","open":19.527,"high":19.6288,"low":19.3828,"close":19.5949,"volume":27492548,"unadjustedVolume":27492548,"change":-0.016962,"changePercent":-0.086,"vwap":19.5245,"label":"Mar 27, 13","changeOverTime":-0.0008617261036722633}]}`,
			want: &Quote{
				Ticker:   "GE",
				Name:     "General Electric Company",
				Exchange: NYSE,
				Last: Last{
					Price:         12.89,
					Time:          time.Unix(1522094417778/1000, 0).In(est),
					Change:        -0.18,
					ChangePercent: -0.01377,
				},
				History: []EOD{
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 19.7391, Close: 19.6118, High: 19.7645, Low: 19.5185, Volume: 32353323},
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 19.527, Close: 19.5949, High: 19.6288, Low: 19.3828, Volume: 27492548},
				},
				Provider: IEXProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			iex := &IEX{
				HTTPClient: &http.Client{},
			}
			got, err := iex.Fetch(tt.args.ticker)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()

}
