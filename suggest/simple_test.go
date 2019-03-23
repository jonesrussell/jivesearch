package suggest

import (
	"reflect"
	"testing"
)

func TestSimpleCompletion(t *testing.T) {
	for _, c := range []struct {
		query  string
		terms  []string
		number int
		want   Results
	}{
		{
			query:  "b",
			terms:  []string{"zebra", "bros", "brad", "bob", "blondie", "brad pitt", "buster"},
			number: 3,
			want: Results{
				Suggestions: []string{"bob", "blondie", "zebra"},
			},
		},
		{
			query:  "q",
			terms:  []string{"zebra", "bros", "brad", "bob", "blondie", "brad pitt", "buster"},
			number: 10,
			want: Results{
				Suggestions: []string{},
			},
		},
	} {
		t.Run(c.query, func(t *testing.T) {
			ms := &Simple{}
			indexExists, err := ms.IndexExists()
			if err != nil {
				t.Fatal(err)
			}

			if !indexExists {
				if err := ms.Setup(); err != nil {
					t.Fatal(err)
				}
			}

			for _, term := range c.terms {
				exists, err := ms.Exists(term)
				if err != nil {
					t.Fatal(err)
				}

				if !exists {
					if err := ms.Insert(term); err != nil {
						t.Fatal(err)
					}
				}

				if err := ms.Increment(term); err != nil {
					t.Fatal(err)
				}

			}

			got, err := ms.Completion(c.query, c.number)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %+v; want %+v", got, c.want)
			}
		})
	}
}
