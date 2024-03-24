package weather

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// OpenWeatherMap retrieves information from the OpenWeatherMap API
type OpenWeatherMap struct {
	HTTPClient *http.Client
	Key        string
}

// OpenWeatherMapProvider is a weather provider
var OpenWeatherMapProvider provider = "OpenWeatherMap"

type owmCurrent struct {
	*OwmInstant
	Coord struct {
		Lon float64 `json:"lon"`
		Lat float64 `json:"lat"`
	} `json:"coord"`
	Base       string `json:"base"`
	Visibility int    `json:"visibility"`
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Cod        int    `json:"cod"`
}

type owmForecast struct {
	Cod     string  `json:"cod"`
	Message float64 `json:"message"`
	Cnt     int     `json:"cnt"`
	List    []struct {
		*OwmInstant
	} `json:"list"`
	City struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Coord struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
		Country string `json:"country"`
	} `json:"city"`
	DtTxt string `json:"dt_txt"` // only forecast data has this
}

// OwmInstant is an OpenWeatherMap day
type OwmInstant struct {
	Instant *Instant
	Dt      int64 `json:"dt"`
	Main    struct {
		Temp      float64 `json:"temp"`
		TempMin   float64 `json:"temp_min"`
		TempMax   float64 `json:"temp_max"`
		Pressure  float64 `json:"pressure"`
		Humidity  int     `json:"humidity"`
		SeaLevel  float64 `json:"sea_level"`
		GrndLevel float64 `json:"grnd_level"`
		TempKf    float64 `json:"temp_kf"`
	} `json:"main"`
	Weather []struct {
		ID          int    `json:"id"`
		Main        string `json:"main"`
		Description string `json:"description"`
		Icon        string `json:"icon"`
	} `json:"weather"`
	Wind struct {
		Speed float64 `json:"speed"`
		Deg   float64 `json:"deg"`
	} `json:"wind"`
	Clouds struct {
		All int `json:"all"`
	} `json:"clouds"`
	Rain struct {
		All float64 `json:"3h"` // rain volume for last 3 hrs
	} `json:"rain"`
	Snow struct {
		All float64 `json:"3h"` // snow volume for last 3 hrs
	} `json:"snow"`
	Sys struct {
		Type    int     `json:"type"`
		ID      int     `json:"id"`
		Message float64 `json:"message"`
		Pod     string  `json:"pod"`
		Country string  `json:"country"`
		Sunrise int     `json:"sunrise"`
		Sunset  int     `json:"sunset"`
	} `json:"sys"`
}

// FetchByCity retrieves weather for a city from the OpenWeatherMap api
func (o *OpenWeatherMap) FetchByCity(city string) (*Weather, error) {
	c := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?APPID=%v&q=%v&units=imperial", o.Key, city)
	f := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?APPID=%v&q=%v&units=imperial", o.Key, city)
	w, err := o.fetchCurrentAndForecast(c, f)
	if err != nil {
		return nil, err
	}

	return w, err
}

// FetchByLatLong retrieves weather for a latitude/longitude location from the OpenWeatherMap api
func (o *OpenWeatherMap) FetchByLatLong(lat, long float64, timeZone string) (*Weather, error) {
	c := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?APPID=%v&lat=%v&lon=%v&units=imperial", o.Key, lat, long)
	f := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?APPID=%v&lat=%v&lon=%v&units=imperial", o.Key, lat, long)
	w, err := o.fetchCurrentAndForecast(c, f)
	if err != nil {
		return nil, err
	}

	w.TimeZone = timeZone
	return w, err
}

// FetchByZip retrieves weather for a zipcode from the OpenWeatherMap api
func (o *OpenWeatherMap) FetchByZip(zip int) (*Weather, error) {
	c := fmt.Sprintf("https://api.openweathermap.org/data/2.5/weather?APPID=%v&zip=%d,us&units=imperial", o.Key, zip)
	f := fmt.Sprintf("https://api.openweathermap.org/data/2.5/forecast?APPID=%v&zip=%d,us&units=imperial", o.Key, zip)

	return o.fetchCurrentAndForecast(c, f)
}

type item struct {
	u   string
	w   interface{}
	bdy io.ReadCloser
	err error
}

func (o *OpenWeatherMap) fetchCurrentAndForecast(currentURL, forecastURL string) (*Weather, error) {
	ch := make(chan item)
	items := []item{
		{
			u: currentURL,
			w: &owmCurrent{},
		},
		{
			u: forecastURL,
			w: &owmForecast{},
		},
	}

	for _, itm := range items {
		go func(itm item, ch chan item) {
			resp, err := o.HTTPClient.Get(itm.u)
			if err != nil {
				itm.err = err
				ch <- itm
				return
			}

			itm.bdy = resp.Body
			ch <- itm
		}(itm, ch)
	}

	w := &Weather{
		Current:  &Instant{},
		Provider: OpenWeatherMapProvider,
	}

	for i := 0; i < len(items); i++ {
		itm := <-ch
		if itm.err != nil {
			return nil, itm.err
		}

		defer itm.bdy.Close()

		switch itm.w.(type) {
		case *owmCurrent:
			t := itm.w.(*owmCurrent)

			if err := json.NewDecoder(itm.bdy).Decode(&t); err != nil {
				return nil, err
			}

			w.City = t.Name // for "Peru" return country of "US"

			d, err := t.instant()
			if err != nil {
				return nil, err
			}

			w.Current = d.Instant
		case *owmForecast:
			f := itm.w.(*owmForecast)
			if err := json.NewDecoder(itm.bdy).Decode(&f); err != nil {
				return nil, err
			}

			forecasts := []*Instant{}
			for _, ff := range f.List {
				fore, err := ff.instant()
				if err != nil {
					return nil, err
				}
				forecasts = append(forecasts, fore.Instant)
			}

			w.Forecast = forecasts
		default:
			fmt.Printf("unknown openweathermap type %T\n", itm.w)
		}
	}

	close(ch)

	return w, nil
}

func (i *OwmInstant) instant() (*OwmInstant, error) {
	i.Instant = &Instant{
		Date:        time.Unix(i.Dt, 0).In(time.UTC),
		Temperature: int(i.Main.Temp + 0.5),
		Low:         int(i.Main.TempMin + 0.5),
		High:        int(i.Main.TempMax + 0.5),
		Wind:        i.Wind.Speed,
		Clouds:      float64(i.Clouds.All),
		Rain:        float64(i.Rain.All),
		Snow:        float64(i.Snow.All),
		Pressure:    float64(i.Main.Pressure),
		Humidity:    float64(i.Main.Humidity),
	}

	for _, cde := range i.Weather {
		if err := i.convertCode(cde.ID); err != nil {
			return nil, err
		}
	}

	return i, nil
}

func (i *OwmInstant) convertCode(id int) error {
	// convert OpenWeatherMap code to a weatherCode
	// http://openweathermap.org/weather-conditions
	switch {
	case id >= 200 && id < 300: // thunderstorm w/ light to heavy rain
		i.Instant.Code = ThunderStorm
	case id >= 300 && id < 399: // light to heavy drizzle
		i.Instant.Code = Rain
	case id >= 500 && id < 599: // light to heavy rain
		i.Instant.Code = Rain
	case id >= 600 && id < 699: // light to heavy snow
		i.Instant.Code = Snow
	case id >= 700 && id < 799: // smoke/haze/fog/tornado, etc
		i.Instant.Code = Extreme
	case id == 800: // clear sky
		i.Instant.Code = Clear
	case id == 801:
		i.Instant.Code = LightClouds
	case id == 802, id == 803:
		i.Instant.Code = ScatteredClouds
	case id == 804:
		i.Instant.Code = OvercastClouds
	case id >= 900 && id < 906:
		// extreme tornado/storm/hurricane/cold/heat...we should probably separate extreme cold & heat from others???
		i.Instant.Code = Extreme
	case id >= 951 && id < 999: // low to high wind
		i.Instant.Code = Windy
	default:
		return fmt.Errorf("unknown OpenWeatherMap code %d", id)
	}

	return nil
}
