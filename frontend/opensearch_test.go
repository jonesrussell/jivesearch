package frontend

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestOpenSearchHandler(t *testing.T) {
	for _, c := range []struct {
		name  string
		brand string
		want  *response
	}{
		{
			"basic", "somegreatname", &response{
				status:   http.StatusOK,
				template: "opensearch",
				data: data{
					Brand: Brand{Name: "somegreatname"},
				},
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Brand: Brand{Name: c.brand},
			}

			req, err := http.NewRequest("GET", "/opensearch.xml", nil)
			if err != nil {
				t.Fatal(err)
			}

			got := f.openSearchHandler(httptest.NewRecorder(), req)

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}
