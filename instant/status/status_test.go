package status

import (
	"reflect"
	"testing"
)

func TestFixDomain(t *testing.T) {
	for _, tt := range []struct {
		name string
		want string
	}{
		{
			name: "example.com",
			want: "example.com",
		},
		{
			name: "www.example.com",
			want: "example.com",
		},
		{
			name: "example",
			want: "example.com",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := FixDomain(tt.name)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
