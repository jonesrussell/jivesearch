package currency

import (
	"reflect"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	for _, tt := range []struct {
		name string
		args provider
		want *Response
	}{
		{
			name: "basic",
			args: ECBProvider,
			want: &Response{
				Base:          USD,
				History:       make(map[string][]*Rate),
				ForexProvider: ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := New()
			got.ForexProvider = tt.args

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
				Base: USD,
				History: map[string][]*Rate{
					JPY.Short: {
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.1,
						},
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.12,
						},
					},
					GBP.Short: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.5,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.6,
						},
					},
				},
				ForexProvider: ECBProvider,
			},
			want: &Response{
				Base: USD,
				History: map[string][]*Rate{
					JPY.Short: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.12,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.1,
						},
					},
					GBP.Short: {
						{
							DateTime: time.Date(2018, 1, 30, 0, 0, 0, 0, time.UTC),
							Rate:     1.5,
						},
						{
							DateTime: time.Date(2018, 1, 31, 0, 0, 0, 0, time.UTC),
							Rate:     1.6,
						},
					},
				},
				ForexProvider: ECBProvider,
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.Sort()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
