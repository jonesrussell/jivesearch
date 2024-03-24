package wikipedia

import (
	"reflect"
	"testing"
)

var shaqRawQuoteJSON = []byte(
	`{"wikibase_item": "Q169452", "language": "en", "title": "Shaquille O'Neal", "source_text": "\n== Quotes ==\n\n* superman rocks\n"}`,
)

var shaqQuotes = []string{"superman rocks"}

var shaqWikiquote = &Wikiquote{
	ID:       "Q169452",
	Language: "en",
	Source:   "\n== Quotes ==\n\n* superman rocks\n",
	Quotes:   shaqQuotes,
}

func TestWikiQuote_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want *Wikiquote
	}{
		{
			"shaq",
			args{shaqRawQuoteJSON},
			shaqWikiquote,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Wikiquote{}

			if err := got.UnmarshalJSON(tt.args.b); err != nil {
				t.Errorf("Wikiquote.UnmarshalJSON() error = %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v; want %+v", got, tt.want)
			}
		})
	}
}
