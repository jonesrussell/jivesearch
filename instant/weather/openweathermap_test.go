package weather

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestOpenWeatherMapFetchByCity(t *testing.T) {
	type args struct {
		city string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		currURL      string
		currResp     string
		forecastURL  string
		forecastResp string
		want         *Weather
	}{
		{
			name:         "Bogota",
			args:         args{"bogota"},
			currURL:      `https://api.openweathermap.org/data/2.5/weather?APPID=myappid&q=bogota&units=imperial`,
			currResp:     `{"coord":{"lon":-111.88,"lat":40.89},"weather":[{"id":802,"main":"Clouds","description":"scattered clouds","icon":"03d"}],"base":"stations","main":{"temp":58.53,"pressure":1014,"humidity":33,"temp_min":55.4,"temp_max":62.6},"visibility":16093,"wind":{"speed":4.7},"clouds":{"all":40},"dt":1522609080,"sys":{"type":1,"id":2802,"message":0.004,"country":"US","sunrise":1522588167,"sunset":1522634001},"id":5771826,"name":"Bogota","cod":200}`,
			forecastURL:  `https://api.openweathermap.org/data/2.5/forecast?APPID=myappid&q=bogota&units=imperial`,
			forecastResp: `{"cod":"200","message":0.0068,"cnt":40,"list":[{"dt":1523469600,"main":{"temp":96.93,"temp_min":83.72,"temp_max":96.93,"pressure":888.01,"sea_level":1025.99,"grnd_level":888.01,"humidity":14,"temp_kf":7.33},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":3.94,"deg":236.5},"sys":{"pod":"d"},"dt_txt":"2018-04-11 18:00:00"},{"dt":1523480400,"main":{"temp":95.32,"temp_min":85.42,"temp_max":95.32,"pressure":886.87,"sea_level":1024.32,"grnd_level":886.87,"humidity":13,"temp_kf":5.5},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":10.76,"deg":233.501},"sys":{"pod":"d"},"dt_txt":"2018-04-11 21:00:00"}],"city":{"id":420000556,"name":"Bogota","coord":{"lat":31.3402,"lon":-110.9361},"country":"CO"}}`,
			want: &Weather{
				City: "Bogota",
				Current: &Instant{
					Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
					Code:        ScatteredClouds,
					Temperature: 59,
					Low:         55,
					High:        63,
					Wind:        4.7,
					Clouds:      40,
					Rain:        0,
					Snow:        0,
					Pressure:    1014,
					Humidity:    33,
				},
				Forecast: []*Instant{
					{
						Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 97,
						Low:         84,
						High:        97,
						Wind:        3.94,
						Pressure:    888.01,
						Humidity:    14,
					},
					{
						Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 95,
						Low:         85,
						High:        95,
						Wind:        10.76,
						Pressure:    886.87,
						Humidity:    13,
					},
				},
				Provider: OpenWeatherMapProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder1 := httpmock.NewStringResponder(200, tt.currResp)
			responder2 := httpmock.NewStringResponder(200, tt.forecastResp)
			httpmock.RegisterResponder("GET", tt.currURL, responder1)
			httpmock.RegisterResponder("GET", tt.forecastURL, responder2)

			o := &OpenWeatherMap{
				HTTPClient: &http.Client{},
				Key:        "myappid",
			}
			got, err := o.FetchByCity(tt.args.city)
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

func TestOpenWeatherMapFetchByLatLong(t *testing.T) {
	type args struct {
		lat      float64
		long     float64
		timezone string
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		currURL      string
		currResp     string
		forecastURL  string
		forecastResp string
		want         *Weather
	}{
		{
			name:         "Centerville, Utah",
			args:         args{40.918, -111.8722, "America/Denver"},
			currURL:      `https://api.openweathermap.org/data/2.5/weather?APPID=myappid&lat=40.918&lon=-111.8722&units=imperial`,
			currResp:     `{"coord":{"lon":-111.88,"lat":40.89},"weather":[{"id":802,"main":"Clouds","description":"scattered clouds","icon":"03d"}],"base":"stations","main":{"temp":58.53,"pressure":1014,"humidity":33,"temp_min":55.4,"temp_max":62.6},"visibility":16093,"wind":{"speed":4.7},"clouds":{"all":40},"dt":1522609080,"sys":{"type":1,"id":2802,"message":0.004,"country":"US","sunrise":1522588167,"sunset":1522634001},"id":5771826,"name":"Bountiful","cod":200}`,
			forecastURL:  `https://api.openweathermap.org/data/2.5/forecast?APPID=myappid&lat=40.918&lon=-111.8722&units=imperial`,
			forecastResp: `{"cod":"200","message":0.0068,"cnt":40,"list":[{"dt":1523469600,"main":{"temp":96.93,"temp_min":83.72,"temp_max":96.93,"pressure":888.01,"sea_level":1025.99,"grnd_level":888.01,"humidity":14,"temp_kf":7.33},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":3.94,"deg":236.5},"sys":{"pod":"d"},"dt_txt":"2018-04-11 18:00:00"},{"dt":1523480400,"main":{"temp":95.32,"temp_min":85.42,"temp_max":95.32,"pressure":886.87,"sea_level":1024.32,"grnd_level":886.87,"humidity":13,"temp_kf":5.5},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":10.76,"deg":233.501},"sys":{"pod":"d"},"dt_txt":"2018-04-11 21:00:00"}],"city":{"id":420000556,"name":"Heroica Nogales","coord":{"lat":31.3402,"lon":-110.9361},"country":"US"}}`,
			want: &Weather{
				City: "Bountiful",
				Current: &Instant{
					Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
					Code:        ScatteredClouds,
					Temperature: 59,
					Low:         55,
					High:        63,
					Wind:        4.7,
					Clouds:      40,
					Rain:        0,
					Snow:        0,
					Pressure:    1014,
					Humidity:    33,
				},
				Forecast: []*Instant{
					{
						Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 97,
						Low:         84,
						High:        97,
						Wind:        3.94,
						Pressure:    888.01,
						Humidity:    14,
					},
					{
						Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 95,
						Low:         85,
						High:        95,
						Wind:        10.76,
						Pressure:    886.87,
						Humidity:    13,
					},
				},
				Provider: OpenWeatherMapProvider,
				TimeZone: "America/Denver",
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder1 := httpmock.NewStringResponder(200, tt.currResp)
			responder2 := httpmock.NewStringResponder(200, tt.forecastResp)
			httpmock.RegisterResponder("GET", tt.currURL, responder1)
			httpmock.RegisterResponder("GET", tt.forecastURL, responder2)

			o := &OpenWeatherMap{
				HTTPClient: &http.Client{},
				Key:        "myappid",
			}
			got, err := o.FetchByLatLong(tt.args.lat, tt.args.long, tt.args.timezone)
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

func TestOpenWeatherMapFetchByZip(t *testing.T) {
	type args struct {
		zip int
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		args
		currURL      string
		currResp     string
		forecastURL  string
		forecastResp string
		want         *Weather
	}{
		{
			name:         "Bountiful, Utah",
			args:         args{84014},
			currURL:      `https://api.openweathermap.org/data/2.5/weather?APPID=myappid&zip=84014,us&units=imperial`,
			currResp:     `{"coord":{"lon":-111.88,"lat":40.89},"weather":[{"id":802,"main":"Clouds","description":"scattered clouds","icon":"03d"}],"base":"stations","main":{"temp":58.53,"pressure":1014,"humidity":33,"temp_min":55.4,"temp_max":62.6},"visibility":16093,"wind":{"speed":4.7},"clouds":{"all":40},"dt":1522609080,"sys":{"type":1,"id":2802,"message":0.004,"country":"US","sunrise":1522588167,"sunset":1522634001},"id":5771826,"name":"Bountiful","cod":200}`,
			forecastURL:  `https://api.openweathermap.org/data/2.5/forecast?APPID=myappid&zip=84014,us&units=imperial`,
			forecastResp: `{"cod":"200","message":0.0068,"cnt":40,"list":[{"dt":1523469600,"main":{"temp":96.93,"temp_min":83.72,"temp_max":96.93,"pressure":888.01,"sea_level":1025.99,"grnd_level":888.01,"humidity":14,"temp_kf":7.33},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":3.94,"deg":236.5},"sys":{"pod":"d"},"dt_txt":"2018-04-11 18:00:00"},{"dt":1523480400,"main":{"temp":95.32,"temp_min":85.42,"temp_max":95.32,"pressure":886.87,"sea_level":1024.32,"grnd_level":886.87,"humidity":13,"temp_kf":5.5},"weather":[{"id":800,"main":"Clear","description":"clear sky","icon":"01d"}],"clouds":{"all":0},"wind":{"speed":10.76,"deg":233.501},"sys":{"pod":"d"},"dt_txt":"2018-04-11 21:00:00"}],"city":{"id":420000556,"name":"Heroica Nogales","coord":{"lat":31.3402,"lon":-110.9361},"country":"US"}}`,
			want: &Weather{
				City: "Bountiful",
				Current: &Instant{
					Date:        time.Date(2018, 4, 1, 18, 58, 0, 0, time.UTC),
					Code:        ScatteredClouds,
					Temperature: 59,
					Low:         55,
					High:        63,
					Wind:        4.7,
					Clouds:      40,
					Rain:        0,
					Snow:        0,
					Pressure:    1014,
					Humidity:    33,
				},
				Forecast: []*Instant{
					{
						Date:        time.Date(2018, 4, 11, 18, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 97,
						Low:         84,
						High:        97,
						Wind:        3.94,
						Pressure:    888.01,
						Humidity:    14,
					},
					{
						Date:        time.Date(2018, 4, 11, 21, 0, 0, 0, time.UTC),
						Code:        Clear,
						Temperature: 95,
						Low:         85,
						High:        95,
						Wind:        10.76,
						Pressure:    886.87,
						Humidity:    13,
					},
				},
				Provider: OpenWeatherMapProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder1 := httpmock.NewStringResponder(200, tt.currResp)
			responder2 := httpmock.NewStringResponder(200, tt.forecastResp)
			httpmock.RegisterResponder("GET", tt.currURL, responder1)
			httpmock.RegisterResponder("GET", tt.forecastURL, responder2)

			o := &OpenWeatherMap{
				HTTPClient: &http.Client{},
				Key:        "myappid",
			}
			got, err := o.FetchByZip(tt.args.zip)
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

func TestConvertCode(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		c    int
		want Description
	}{
		{200, ThunderStorm},
		{300, Rain},
		{500, Rain},
		{600, Snow},
		{700, Extreme},
		{800, Clear},
		{801, LightClouds},
		{804, OvercastClouds},
		{900, Extreme},
		{951, Windy},
	} {
		t.Run(fmt.Sprint(tt.c), func(t *testing.T) {
			r := &OwmInstant{
				Instant: &Instant{},
			}

			if err := r.convertCode(tt.c); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(r.Instant.Code, tt.want) {
				t.Errorf("got %+v, want %+v", r.Instant.Code, tt.want)
			}
		})
	}

	httpmock.Reset()
}
