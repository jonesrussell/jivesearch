package breach

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	type args struct {
		account string
		provider
	}

	for _, tt := range []struct {
		name string
		args
		want *Response
	}{
		{
			name: "basic",
			args: args{"someone@example.com", HaveIBeenPwnedProvider},
			want: &Response{
				Account:  "someone@example.com",
				Provider: HaveIBeenPwnedProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := New(tt.args.account, tt.args.provider)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSort(t *testing.T) {
	for _, tt := range []struct {
		name string
		args *Response
		want *Response
	}{
		{
			name: "basic",
			args: &Response{
				Breaches: []Breach{
					{Date: time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC)},
					{Date: time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC)},
					{Date: time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC)},
				},
			},
			want: &Response{
				Breaches: []Breach{
					{Date: time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC)},
					{Date: time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC)},
					{Date: time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC)},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt.args.Sort()

			if !reflect.DeepEqual(tt.args, tt.want) {
				t.Errorf("got %+v, want %+v", tt.args, tt.want)
			}
		})
	}
}
