package stock

import (
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestSortHistorical(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input *Quote
		want  *Quote
	}{
		{
			name: "sorted",
			input: &Quote{
				History: []EOD{
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
				},
			},
			want: &Quote{
				History: []EOD{
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
				},
			},
		},
		{
			name: "backwards",
			input: &Quote{
				History: []EOD{
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 19.527, Close: 19.5949, High: 19.6288, Low: 19.3828, Volume: 27492548},
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 19.7391, Close: 19.6118, High: 19.7645, Low: 19.5185, Volume: 32353323},
				},
			},
			want: &Quote{
				History: []EOD{
					{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 19.7391, Close: 19.6118, High: 19.7645, Low: 19.5185, Volume: 32353323},
					{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 19.527, Close: 19.5949, High: 19.6288, Low: 19.3828, Volume: 27492548},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.SortHistorical()

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()

}
