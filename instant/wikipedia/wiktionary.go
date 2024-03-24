package wikipedia

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// Wiktionary holds the structure for a word and it's definition(s)
type Wiktionary struct {
	Title    string `json:"title"`
	Language string `json:"language,omitempty"`
	Source   string `json:"source_text,omitempty"` // "text" isn't parseable
	// Etymology     string // origin of the word...not implemented yet
	// Pronunciation string // not implemented yet
	Definitions []*Definition `json:"definitions,omitempty"`
}

// Definition is a single definition and synonyms
type Definition struct { // noun, verb, pronoun, adjective, adverb, preposition, conjunction,  interjection, determiner
	Part     string    `json:"part,omitempty"`
	Meaning  string    `json:"meaning,omitempty"`
	Synonyms []Synonym `json:"synonyms,omitempty"`
}

// Synonym is a Wiktionary link to another word
type Synonym struct {
	Language string `json:"language,omitempty"`
	Word     string `json:"word,omitempty"`
}

var parts = []string{"noun", "proper noun", "verb", "pronoun", "adjective", "adverb", "preposition", "conjunction", "interjection", "determiner"}

var reQuadEq = regexp.MustCompile(`(?m)^====(.+?)====$`)
var reTripleEq = regexp.MustCompile(`(?m)^===(.+?)===$`)
var reDoubleEq = regexp.MustCompile(`(?m)^==(.+?)==$`)
var reSingleEq = regexp.MustCompile(`={1}`)

// change header tags back to equal signs so we can use [^=] matching below
var reH1 = regexp.MustCompile(`(?m)^<h1>(.+?)</h1>$`)
var reH2 = regexp.MustCompile(`(?m)^<h2>(.+?)</h2>`)
var reH3 = regexp.MustCompile(`(?m)^<h3>(.+?)</h3>`)

var h1 = regexp.MustCompile(`(?m)^==(.*?)==([^=]*)`)
var h2 = regexp.MustCompile(`(?m)^===(.*?)===([^=]*)`)
var h3 = regexp.MustCompile(`(?m)^====(.*?)====([^=]*)`)
var orderedList = regexp.MustCompile(`[\n\r]?#(.+)`)
var wrdRegex = regexp.MustCompile(`{{([^{}|]*?)\|([^{}|]*?)\|([^{}|]*?)}}`)

// UnmarshalJSON extracts the raw info needed from the source_text
func (w *Wiktionary) UnmarshalJSON(data []byte) error {
	// copy the fields of Wikipedia but not the
	// methods so we don't recursively call UnmarshalJSON
	type Alias Wiktionary
	a := &struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	}

	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	// change the equal signs so they get out of the way of regex
	w.Source = reQuadEq.ReplaceAllString(w.Source, "<h3>$1</h3>")   // Subsubheading
	w.Source = reTripleEq.ReplaceAllString(w.Source, "<h2>$1</h2>") // Subheading
	w.Source = reDoubleEq.ReplaceAllString(w.Source, "<h1>$1</h1>") // Heading
	w.Source = reSingleEq.ReplaceAllString(w.Source, "[equals]")
	w.Source = reH1.ReplaceAllString(w.Source, "==$1==") // Change heading back so we can use [^=]

	for _, m := range h1.FindAllStringSubmatch(w.Source, -1) {
		if len(m) < 2 {
			continue
		}

		// change <h2> back to "==="
		content := reH2.ReplaceAllString(strings.TrimSpace(m[2]), "===$1===") // Subheading

		// get the <h2> subheadings
		for _, mm := range h2.FindAllStringSubmatch(content, -1) {
			if len(mm) < 2 {
				continue
			}

			subHeading := strings.ToLower(strings.TrimSpace(mm[1]))

			for _, p := range parts {
				if subHeading == p {
					definition := &Definition{
						Part: p,
					}

					// find noun, verb, etc definitions
					for _, item := range orderedList.FindAllStringSubmatch(strings.TrimSpace(mm[2]), -1) {
						if len(item) < 2 {
							continue
						}

						definition.Meaning = reBraces.ReplaceAllString(item[1], "")
						definition.Meaning = reBrackets.ReplaceAllString(definition.Meaning, "$1")
						definition.Meaning = strings.TrimPrefix(definition.Meaning, "* ")
						definition.Meaning = strings.TrimSpace(definition.Meaning)
						definition.Meaning = bluemonday.StrictPolicy().Sanitize(definition.Meaning)

						break // only need the first def?
					}

					// synonyms
					c := reH3.ReplaceAllString(strings.TrimSpace(mm[2]), "====$1====")
					for _, sy := range h3.FindAllStringSubmatch(c, -1) {
						if len(sy) < 2 {
							continue
						}

						if strings.TrimSpace(sy[1]) != "Synonyms" {
							continue
						}

						for _, s := range strings.Split(sy[2], ",") {
							for _, w := range wrdRegex.FindAllStringSubmatch(strings.TrimSpace(s), -1) {
								if len(w) < 3 || w[1] != "l" {
									continue
								}

								sy := Synonym{
									Language: w[2],
									Word:     w[3],
								}

								definition.Synonyms = append(definition.Synonyms, sy)
							}
						}
					}

					w.Definitions = append(w.Definitions, definition)
				}
			}
		}

		// Grab the first language for now.
		// When we get the Eyymoloyg we might need the other languages
		break
	}

	return nil
}
