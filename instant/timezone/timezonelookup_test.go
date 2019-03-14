package timezone

import (
	"reflect"
	"testing"

	timezone "github.com/evanoberholster/timezoneLookup"
)

func TestTZLookupFetch(t *testing.T) {
	type args struct {
		lat float64
		lon float64
	}

	for _, tt := range []struct {
		name string
		args
		want string
	}{
		{
			name: "basic",
			args: args{-33.8667, 151.2},
			want: "Australia/Sydney",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockTZLookupInterface{}
			tz := &TZLookup{m}

			got, err := tz.Fetch(tt.args.lat, tt.args.lon)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %q; want %q", got, tt.want)
			}
		})
	}
}

type mockTZLookupInterface struct{}

func (m *mockTZLookupInterface) CreateTimezones(jsonFilename string) error {
	return nil
}
func (m *mockTZLookupInterface) LoadTimezones() error {
	return nil
}
func (m *mockTZLookupInterface) Query(q timezone.Coord) (string, error) {
	sydney := timezone.Coord{
		Lat: 151.2,
		Lon: -33.8667,
	}

	if reflect.DeepEqual(q, sydney) {
		return "Australia/Sydney", nil
	}

	return "America/Denver", nil
}
func (m *mockTZLookupInterface) Close() {}
