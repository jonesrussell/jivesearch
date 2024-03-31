package main

import (
	"reflect"
	"testing"

	"github.com/jonesrussell/jivesearch/instant/wikipedia"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

func TestSetup(t *testing.T) {
	// Continue with your test logic...
	// Instantiate Frontend and set its TemplateParser field
	v := viper.New()
	s := setup(v)

	want := ":8000"
	if s.Addr != want {
		t.Fatalf("got %q; want %q", s.Addr, want)
	}

	if s.Handler == nil {
		t.Fatalf("got nil handler")
	}
}

func TestLanguages(t *testing.T) {
	type want struct {
		supported   []language.Tag
		unsupported int
	}

	tests := []struct {
		name string
		args []string
		want
	}{
		{"default", []string{}, want{[]language.Tag{}, 0}},
		{"en", []string{"en"}, want{[]language.Tag{language.English}, 0}},
		{"custom", []string{"fr", "de"}, want{[]language.Tag{language.French, language.German}, 0}},
		{"unsupported", []string{"en", "def"}, want{[]language.Tag{language.English}, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.SetDefault("languages", tt.args)

			supported, unsupported := languages(v)
			if len(unsupported) != tt.want.unsupported {
				t.Errorf("got %d unsupported, want %d", len(unsupported), tt.want.unsupported)
			}

			switch len(tt.args) {
			case 0:
				if len(supported) != len(wikipedia.Available) {
					t.Errorf("got %+v, want %+v", supported, tt.want.supported)
				}
			default:
				if !reflect.DeepEqual(supported, tt.want.supported) {
					t.Errorf("got %+v, want %+v", supported, tt.want.supported)
				}
			}
		})
	}
}
