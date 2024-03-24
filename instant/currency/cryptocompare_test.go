package currency

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestCryptoCompareFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		u    map[string]string
		want *Response
	}{
		{
			name: "basic",
			u: map[string]string{
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=BTC&limit=90&tsym=USD`:  `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":6323.81,"high":6478.07,"low":6217.33,"open":6274.22,"volumefrom":71357.92,"volumeto":454679037.36},{"time":1534464000,"close":6591.16,"high":6594.72,"low":6300.45,"open":6323.81,"volumefrom":73383.05,"volumeto":477089455.56},{"time":1534550400,"close":6574.48,"high":6605.47,"low":6560.74,"open":6591.18,"volumefrom":1878.99,"volumeto":12415731.08}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=DOGE&limit=90&tsym=USD`: `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":0.002282,"high":0.00245,"low":0.00223,"open":0.002364,"volumefrom":14108560.77,"volumeto":33102.64},{"time":1534464000,"close":0.00254,"high":0.00258,"low":0.002276,"open":0.002276,"volumefrom":10230776.5,"volumeto":24899.45},{"time":1534550400,"close":0.00255,"high":0.00255,"low":0.002525,"open":0.002537,"volumefrom":142647.98,"volumeto":361.79}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=ETH&limit=90&tsym=USD`:  `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":286.8,"high":298.84,"low":278.02,"open":281.24,"volumefrom":533236.46,"volumeto":153825432.04},{"time":1534464000,"close":317.57,"high":318.31,"low":285.34,"open":286.8,"volumefrom":733901.78,"volumeto":221725330.97},{"time":1534550400,"close":314.48,"high":319.61,"low":314.38,"open":317.57,"volumefrom":12636.85,"volumeto":4007565.38}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=LTC&limit=90&tsym=USD`:  `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":55.39,"high":57.63,"low":54.15,"open":54.41,"volumefrom":282293.19,"volumeto":15751740.76},{"time":1534464000,"close":61.79,"high":61.83,"low":55.14,"open":55.39,"volumefrom":429438.87,"volumeto":25170257.08},{"time":1534550400,"close":61.02,"high":61.99,"low":61.02,"open":61.79,"volumefrom":13955.3,"volumeto":857651.4}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=XMR&limit=90&tsym=USD`:  `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":91.28,"high":92.83,"low":86.77,"open":89.36,"volumefrom":19945.08,"volumeto":1806894.23},{"time":1534464000,"close":99.45,"high":101.6,"low":90.61,"open":91.28,"volumefrom":27235.27,"volumeto":2638831.52},{"time":1534550400,"close":100.07,"high":100.07,"low":99.22,"open":99.45,"volumefrom":181.99,"volumeto":18066.52}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
				`https://min-api.cryptocompare.com/data/histoday?extraParams=testagent&fsym=XRP&limit=90&tsym=USD`:  `{"Response":"Success","Type":100,"Aggregated":false,"Data":[{"time":1534377600,"close":0.2916,"high":0.3015,"low":0.2758,"open":0.2803,"volumefrom":64078392.97,"volumeto":18628784.76},{"time":1534464000,"close":0.3672,"high":0.3736,"low":0.291,"open":0.2916,"volumefrom":129218463.39,"volumeto":42632206.54},{"time":1534550400,"close":0.3539,"high":0.3705,"low":0.3539,"open":0.3672,"volumefrom":5336547.05,"volumeto":1944820.66}],"TimeTo":1534550400,"TimeFrom":1534377600,"FirstValueInArray":true,"ConversionType":{"type":"direct","conversionSymbol":""}}`,
			},
			want: &Response{
				Base: USD,
				History: map[string][]*Rate{
					BTC.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     6323.81,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     6591.16,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     6574.48,
						},
					},
					DOGE.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     0.002282,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     0.00254,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     0.00255,
						},
					},
					ETH.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     286.8,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     317.57,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     314.48,
						},
					},
					LTC.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     55.39,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     61.79,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     61.02,
						},
					},
					XMR.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     91.28,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     99.45,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     100.07,
						},
					},
					XRP.Short: {
						{
							DateTime: time.Date(2018, 8, 16, 0, 0, 0, 0, time.UTC),
							Rate:     0.2916,
						},
						{
							DateTime: time.Date(2018, 8, 17, 0, 0, 0, 0, time.UTC),
							Rate:     0.3672,
						},
						{
							DateTime: time.Date(2018, 8, 18, 0, 0, 0, 0, time.UTC),
							Rate:     0.3539,
						},
					},
				},
				CryptoProvider: CryptoCompareProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.u {
				responder := httpmock.NewStringResponder(200, v)
				httpmock.RegisterResponder("GET", k, responder)
			}

			cc := &CryptoCompare{
				Client:    &http.Client{},
				UserAgent: "testagent",
			}
			got, err := cc.Fetch()
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
