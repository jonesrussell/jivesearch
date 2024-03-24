// Package wikipedia fetches Wikipedia articles
package wikipedia

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// Wikiquote holds the summary text of an article
// another option is xml: https://dumps.wikimedia.org/enwikiquote/20180201/enwikiquote-20180201-pages-articles-multistream.xml.bz2
type Wikiquote struct {
	ID       string   `json:"wikibase_item,omitempty"`
	Language string   `json:"language,omitempty"`
	Source   string   `json:"source_text,omitempty"` // "text" isn't parseable
	Quotes   []string `json:"quotes,omitempty"`
}

// Another option is to convert wiky to Go then work w/ the HTML
// but there are PCRE regex patterns that won't work with Go's RE2.
// https://github.com/lahdekorpi/Wiky.php
var reRefTags = regexp.MustCompile(`<ref>.*?</ref>`)                        // "<ref>https://www.example.com</ref>" => ""
var reWikiLinks = regexp.MustCompile(`(.*)(\[{2})(.*?)\|(.*?)(\]{2})(.*?)`) // "a link to [[w:Pasta|Pasta]] here" => "a link to Pasta here"
var reBraces = regexp.MustCompile(`{{.*?}}`)                                // "{{citation}}" => ""
var reBrackets = regexp.MustCompile(`\[\[(.*?)\]\]`)                        // "[[Gratitude]]" => "Gratitude"
var sanitizer = bluemonday.StrictPolicy()

// UnmarshalJSON extracts the raw quotes from the source_text
func (w *Wikiquote) UnmarshalJSON(data []byte) error {
	// copy the fields of Wikiquote but not the
	// methods so we don't recursively call UnmarshalJSON
	type Alias Wikiquote
	a := &struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	}

	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	// The following is wikitext.
	// == indicates a new section (Quotes, Song lyrics, etc...)
	// === indicates a new subsection (eg the albulm of the song lyrics)
	// \n\n* is a new quote
	// \n** indicates a source
	// '''some text''' indicates bold font
	// Wikitext cheat sheet: https://en.wikipedia.org/wiki/Help:Cheatsheet
	// not sure how to capture song lyrics (albums start with "===")

	// change the equal signs so they get out of the way of regex
	w.Source = reQuadEq.ReplaceAllString(w.Source, "<h3>$1</h3>")   // Subsubheading
	w.Source = reTripleEq.ReplaceAllString(w.Source, "<h2>$1</h2>") // Subheading
	w.Source = reDoubleEq.ReplaceAllString(w.Source, "<h1>$1</h1>") // Heading
	w.Source = reSingleEq.ReplaceAllString(w.Source, "[equals]")
	w.Source = reH2.ReplaceAllString(w.Source, "===$1===") // Change heading back so we can use [^=]
	w.Source = reH1.ReplaceAllString(w.Source, "==$1==")   // Change heading back so we can use [^=]

	for _, m := range h1.FindAllStringSubmatch(w.Source, -1) {
		if len(m) < 2 {
			continue
		}
		section := strings.ToLower(strings.TrimSpace(m[1]))
		if section == "quotes" || section == "sourced" { // any more sections we want???
			for _, q := range strings.Split(m[2], "\n") {
				if strings.HasPrefix(q, "* ") {
					// Remove wikitext formatting. I couldn't find a good
					// library to convert wikitext to html or some other format.
					q = strings.TrimPrefix(q, "* ")
					q = strings.Replace(q, `'''`, "", -1)
					q = reRefTags.ReplaceAllString(q, "")
					q = reWikiLinks.ReplaceAllString(q, `$1$4$6`)
					q = reBraces.ReplaceAllString(q, "")
					q = reBrackets.ReplaceAllString(q, "$1")
					q = sanitizer.Sanitize(q) // run this AFTER we strip <ref>link</ref> section

					w.Quotes = append(w.Quotes, q)
				}
			}
		}
	}

	return nil
}
