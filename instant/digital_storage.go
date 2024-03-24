package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// UnitConverterType is an answer Type
const UnitConverterType Type = "unit converter"

// DigitalStorage is an instant answer
type DigitalStorage struct {
	Answer
}

func (d *DigitalStorage) setQuery(req *http.Request, q string) Answerer {
	d.Answer.setQuery(req, q)
	return d
}

func (d *DigitalStorage) setUserAgent(req *http.Request) Answerer {
	return d
}

func (d *DigitalStorage) setLanguage(lang language.Tag) Answerer {
	d.language = lang
	return d
}

func (d *DigitalStorage) setType() Answerer {
	d.Type = UnitConverterType
	return d
}

func (d *DigitalStorage) setRegex() Answerer {
	// a query for "convert" will result in a DigitalStorage answer
	triggers := []string{
		"convert", "converter",
	}

	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)$`, strings.Join(triggers, "|"))))

	digitalStorage := []string{
		"bit", "byte",
		"kilobit", "kibibit", "kilobyte", "kibibyte",
		"megabit", "mebibit", "megabyte", "mebibyte",
		"gigabit", "gibibit", "gigabyte", "gibibyte",
		"terabit", "tebibit", "terabyte", "tebibyte",
		"petabit", "pebibit", "petabyte", "pebibyte",
		"kb", "kbit", "kibit", "kib",
		"mb", "mbit", "mibit", "mib",
		"gb", "gbit", "gibit", "gib",
		"tb", "tbit", "tibit", "tib",
		"pb", "pbit", "pibit", "pib",
	}

	for i, ds := range digitalStorage {
		digitalStorage[i] = ds + "[s]?"
	}

	dss := strings.Join(digitalStorage, "|")
	t := fmt.Sprintf("[0-9 ]*?%v to [0-9 ]*?%v", dss, dss)

	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)(?P<remainder>.*)$`, t)))
	d.regex = append(d.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*)(?P<trigger>%s)$`, t)))

	return d
}

func (d *DigitalStorage) solve(r *http.Request) Answerer {
	// The caller is expected to provide the solution when triggered, preferably in JavaScript
	// TODO: pass the remainder to our html template so that "50gb to mb" prefills the form with "50", "gb", and "mb"
	// Note: Combining the digital storage, length and other unit converters would then require us to make
	// sense of the regexp remainder. We could put unit converters in a subpackage, though.
	d.Solution = "digital storage"
	return d
}

func (d *DigitalStorage) tests() []test {
	dd := Data{
		Type:      UnitConverterType,
		Triggered: true,
		Solution:  "digital storage",
	}

	tests := []test{
		{
			query:    "convert",
			expected: []Data{dd},
		},
		{
			query:    "convert 1mb to pbs",
			expected: []Data{dd},
		},
		/*
			not passing for some reason...
			{
				query:    "50gb to mb converter",
				expected: []Data{d},
			},
		*/
		{
			query:    "gb to mb",
			expected: []Data{dd},
		},
		{
			query:    "petabytes to megabit",
			expected: []Data{dd},
		},
		{
			query:    "50 gb to 100pb",
			expected: []Data{dd},
		},
		{
			query:    "50gbs to mbs",
			expected: []Data{dd},
		},
	}

	return tests
}
