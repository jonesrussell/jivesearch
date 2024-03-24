package bangs

import (
	"reflect"
	"testing"
)

func TestSimpleSuggest(t *testing.T) {
	bngs := &Bangs{
		Bangs: []Bang{
			{Triggers: []string{"g", "google"}},
			{Triggers: []string{"gfr"}},
			{Triggers: []string{"g", "google"}},
			{Triggers: []string{"g", "google"}},
		},
	}

	for _, c := range []struct {
		query  string
		terms  []string
		number int
		want   Results
	}{
		{
			query:  "!g",
			terms:  []string{"g", "goog", "gfr", "bing", "yf"},
			number: 3,
			want: Results{
				Suggestions: []Suggestion{
					{Trigger: "g"},
					{Trigger: "gfr"},
					{Trigger: "google"},
				},
			},
		},
		{
			query:  "!q",
			terms:  []string{"g", "goog", "gfr", "bing", "yf"},
			number: 10,
			want:   Results{},
		},
	} {
		t.Run(c.query, func(t *testing.T) {
			ms := &Simple{}
			indexExists, err := ms.IndexExists()
			if err != nil || indexExists {
				t.Fatal(err)
			}

			if err := ms.DeleteIndex(); err != nil {
				t.Fatal(err)
			}

			if err := ms.Setup(bngs.Bangs); err != nil {
				t.Fatal(err)
			}

			got, err := ms.SuggestResults(c.query, c.number)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}

}
