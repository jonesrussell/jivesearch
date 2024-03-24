package location

import (
	"net"
	"reflect"
	"testing"

	geoip2 "github.com/oschwald/geoip2-golang"
)

func TestFetch(t *testing.T) {
	type args struct {
		ip net.IP
	}

	type want struct {
		city map[string]string
	}

	for _, tt := range []struct {
		name string
		args
		want
	}{
		{
			name: "127.0.0.1",
			args: args{net.ParseIP("127.0.0.1")},
			want: want{
				city: map[string]string{"en": "Centerville", "zh-CN": "森特维尔"},
			},
		},
		{
			name: "179.131.73.22",
			args: args{net.ParseIP("179.131.73.22")},
			want: want{
				city: map[string]string{"en": "Someville"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			open = func(loc string) (maxMinder, error) {
				return &mockMinder{}, nil
			}

			mm := &MaxMind{}
			got, err := mm.Fetch(tt.args.ip)
			if err != nil {
				t.Fatal(err)
			}

			want := &City{}
			want.City.Names = tt.want.city

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

type mockMinder struct{}

func (m *mockMinder) Close() error {
	return nil
}

func (m *mockMinder) City(ipAddress net.IP) (*geoip2.City, error) {
	ip := ipAddress.String()

	c := &geoip2.City{}

	switch ip {
	case "127.0.0.1":
		c.City.Names = map[string]string{"en": "Centerville", "zh-CN": "森特维尔"}
	case "179.131.73.22":
		c.City.Names = map[string]string{"en": "Someville"}

	}

	return c, nil
}
