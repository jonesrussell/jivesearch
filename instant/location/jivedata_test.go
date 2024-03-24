package location

import (
	"net"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestJiveDataFetch(t *testing.T) {
	type args struct {
		ip net.IP
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		raw  string
		args
	}{
		{
			name: "127.0.0.1",
			raw:  `{"ip":["127.0.0.1"],"hostname":["localhost"],"isp":"","asn_number":0,"asn":"","city":{"geoname_id":0,"names":null},"continent":{"code":"","geoname_id":0,"names":null},"country":{"geoname_id":0,"is_eu":false,"iso_code":"","names":null},"location":{"accuracy_radius":0,"latitude":0,"longitude":0,"metro_code":0,"time_zone":""},"postal":{"code":""},"registered_country":{"geoname_id":0,"is_eu":false,"iso_code":"","names":null},"represented_country":{"geoname_id":0,"is_eu":false,"iso_code":"","names":null,"type":""},"subdivisions":null,"traits":{"is_anonymous_proxy":false,"is_satellite_provider":false}}`,
			args: args{net.ParseIP("127.0.0.1")},
		},
		{
			name: "179.131.73.22",
			raw:  `{"ip":["179.131.73.22"],"hostname":["179.131.73.22"],"isp":"Telefonica Data S.A.","asn_number":11419,"asn":"Telefonica Data S.A.","city":{"geoname_id":0,"names":null},"continent":{"code":"SA","geoname_id":6255150,"names":{"de":"Südamerika","en":"South America","es":"Sudamérica","fr":"Amérique du Sud","ja":"南アメリカ","pt-BR":"América do Sul","ru":"Южная Америка","zh-CN":"南美洲"}},"country":{"geoname_id":3469034,"is_eu":false,"iso_code":"BR","names":{"de":"Brasilien","en":"Brazil","es":"Brasil","fr":"Brésil","ja":"ブラジル連邦共和国","pt-BR":"Brasil","ru":"Бразилия","zh-CN":"巴西"}},"location":{"accuracy_radius":1000,"latitude":-22.8305,"longitude":-43.2192,"metro_code":0,"time_zone":""},"postal":{"code":""},"registered_country":{"geoname_id":3469034,"is_eu":false,"iso_code":"BR","names":{"de":"Brasilien","en":"Brazil","es":"Brasil","fr":"Brésil","ja":"ブラジル連邦共和国","pt-BR":"Brasil","ru":"Бразилия","zh-CN":"巴西"}},"represented_country":{"geoname_id":0,"is_eu":false,"iso_code":"","names":null,"type":""},"subdivisions":null,"traits":{"is_anonymous_proxy":false,"is_satellite_provider":false}}`,
			args: args{net.ParseIP("179.131.73.22")},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			u, err := url.Parse("https://jivedata.com/geolocation")
			if err != nil {
				t.Fatal(err)
			}

			q := u.Query()
			q.Set("key", j.Key)
			q.Set("ip", tt.ip.String())
			u.RawQuery = q.Encode()

			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", u.String(), responder)

			got, err := j.Fetch(tt.args.ip)
			if err != nil {
				t.Fatal(err)
			}

			want := cityWant(tt.args.ip.String())

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}

func cityWant(ip string) *City {
	c := &City{}

	switch ip {
	case "127.0.0.1":
	case "179.131.73.22":
		c.Continent.Code = "SA"
		c.Continent.GeoNameID = 6255150
		c.Continent.Names = map[string]string{
			"pt-BR": "América do Sul",
			"ru":    "Южная Америка",
			"zh-CN": "南美洲",
			"de":    "Südamerika",
			"en":    "South America",
			"es":    "Sudamérica",
			"fr":    "Amérique du Sud",
			"ja":    "南アメリカ",
		}

		c.Country.GeoNameID = 3469034
		c.Country.IsInEuropeanUnion = false
		c.Country.IsoCode = "BR"
		c.Country.Names = map[string]string{
			"ru":    "Бразилия",
			"zh-CN": "巴西",
			"de":    "Brasilien",
			"en":    "Brazil",
			"es":    "Brasil",
			"fr":    "Brésil",
			"ja":    "ブラジル連邦共和国",
			"pt-BR": "Brasil",
		}

		c.Location.AccuracyRadius = 1000
		c.Location.Latitude = -22.8305
		c.Location.Longitude = -43.2192
		c.Location.MetroCode = 0

		c.RegisteredCountry.GeoNameID = 3469034
		c.RegisteredCountry.IsInEuropeanUnion = false
		c.RegisteredCountry.IsoCode = "BR"
		c.RegisteredCountry.Names = map[string]string{
			"ru":    "Бразилия",
			"zh-CN": "巴西",
			"de":    "Brasilien",
			"en":    "Brazil",
			"es":    "Brasil",
			"fr":    "Brésil",
			"ja":    "ブラジル連邦共和国",
			"pt-BR": "Brasil",
		}
	}

	return c
}
