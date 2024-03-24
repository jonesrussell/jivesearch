// Package suggest handles AutoComplete and Phrase Suggester (Did you mean?) queries
package suggest

import (
	"bufio"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/afero"
)

// Suggester outlines methods to fetch & store Autocomplete & PhraseSuggester results
type Suggester interface {
	IndexExists() (bool, error)
	Setup() error
	Exists(q string) (bool, error)
	Insert(q string) error
	Increment(q string) error
	Completion(q string, size int) (Results, error)
	//phrase(q string) Results //  TODO: "Did you mean?"
}

// Results are the results of an autocomplete query
type Results struct { // remember top-level arrays = no-no in javascript/json
	Suggestions []string `json:"suggestions"`
}

var appFs = afero.NewOsFs()

// NewNaughty loads our naughty list
func NewNaughty(fh string) error {
	file, err := appFs.Open(fh)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		wrd := scanner.Text()
		if strings.HasPrefix(wrd, "#") { // skip the first line
			continue
		}
		naughty[strings.ToLower(wrd)] = struct{}{}
	}

	if len(naughty) == 0 {
		return fmt.Errorf("no naughty words")
	}

	return scanner.Err()
}

var naughty = make(map[string]struct{})

// Naughty indicates if a word or phrase is NSFW
func Naughty(s string) bool {
	for k := range naughty {
		if strings.Contains(strings.ToLower(s), k) {
			return true
		}
	}

	return false
}
