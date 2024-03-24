package population

import (
	"reflect"
	"testing"
	"time"
)

func TestSort(t *testing.T) {
	for _, tt := range []struct {
		name string
		args *Response
		want *Response
	}{
		{
			name: "basic",
			args: &Response{
				History: []Instant{
					{time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC), 18},
					{time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC), 4},
					{time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC), 2},
				},
			},
			want: &Response{
				History: []Instant{
					{time.Date(1994, 12, 31, 0, 0, 0, 0, time.UTC), 4},
					{time.Date(2003, 12, 31, 0, 0, 0, 0, time.UTC), 2},
					{time.Date(2017, 12, 31, 0, 0, 0, 0, time.UTC), 18},
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
