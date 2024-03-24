package wikipedia

import (
	"reflect"
	"testing"
)

var guitarRawJSON = []byte(`{
	"source_text": "==English==\n\n===Noun===\n# musical instrument\n\n====Synonyms====\n* {{l|en|axe}}",
	"wiki": "enwiktionary",
	"language": "en",
	"title": "guitar",
	"text": "this part isn't important",
	"popularity_score": 6.559202981352e-6
}`)

var guitarWiktionary = &Wiktionary{
	Title:    "guitar",
	Language: "en",
	Definitions: []*Definition{
		{
			Part:    "noun",
			Meaning: "musical instrument",
			Synonyms: []Synonym{
				{Language: "en", Word: "axe"},
			},
		},
	},
}

func TestWiktionary_UnmarshalJSON(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want *Wiktionary
	}{
		{
			"guitar",
			args{guitarRawJSON},
			guitarWiktionary,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &Wiktionary{}

			if err := got.UnmarshalJSON(tt.args.b); err != nil {
				t.Errorf("Wikiquote.UnmarshalJSON() error = %v", err)
			}

			got.Source = ""

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %+v; want %+v", got, tt.want)
			}
		})
	}
}
