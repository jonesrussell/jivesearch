package wikipedia

import (
	"bufio"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/afero"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"golang.org/x/text/language"
)

// CirrusURL is the url for the cirrus wikipedia files
var CirrusURL, _ = url.Parse("https://dumps.wikimedia.org/other/cirrussearch/current/")

// WikiDataURL comes from a different url (smaller file)...cirrus link is formatted differently.
var WikiDataURL, _ = url.Parse("https://dumps.wikimedia.org/wikidatawiki/entities/latest-all.json.bz2")

// FileType is a type of Wikipedia file
type FileType string

const (
	// WikidataFT is a Wikidata file type
	WikidataFT FileType = "wikidata"
	// WikipediaFT is a Wikipedia file type
	WikipediaFT FileType = "wikipedia"
	// WikiquoteFT is a Wikiquote file type
	WikiquoteFT FileType = "wikiquote"
	// WiktionaryFT is a Wiktionary file type
	WiktionaryFT FileType = "wiktionary"
)

// File is a wikipedia/wikidata dump file
type File struct {
	URL      *url.URL
	language language.Tag
	Base     string
	Dir      string
	ABS      string
	Type     FileType
	dumper
}

// allows for filesystem mock in tests
var fs = afero.NewOsFs()

// dumper outlines methods to dump raw files to a database
type dumper interface {
	Dump(ft FileType, lang language.Tag, rows chan interface{}) error
}

// NewFile returns a new file and sets the URL and Base.
func NewFile(u *url.URL, ft FileType, l language.Tag) *File {
	return &File{
		URL:      u,
		language: l,
		Type:     ft,
		Base:     path.Base(u.Path),
	}
}

// SetDumper sets the Dumper for a file
func (f *File) SetDumper(d dumper) *File {
	f.dumper = d
	return f
}

// SetABS sets the absolute path for a file
func (f *File) SetABS(dir string) *File {
	f.ABS = filepath.Join(dir, f.Base)
	return f
}

// Download downloads a wikipedia/wikidata dump file
func (f *File) Download() error {
	resp, err := http.Get(f.URL.String())
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	out, err := fs.Create(f.ABS)
	if err != nil {
		return err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

const maxCapacity = 10000 * 1024 // seems to work

// Parse parses a wikipedia/wikidata dump file and sends it to Dumper
func (f *File) Parse(truncate int) error {
	ff, err := fs.Open(f.ABS)
	if err != nil {
		return err
	}
	defer ff.Close()

	ext := filepath.Ext(f.ABS)
	var scanner *bufio.Scanner

	switch ext {
	case ".bz2": // wikidata
		rdr := bzip2.NewReader(ff)
		scanner = bufio.NewScanner(rdr)
	case ".gz":
		rdr, err := gzip.NewReader(ff)
		if err != nil {
			return err
		}
		defer rdr.Close()
		scanner = bufio.NewScanner(rdr)
	default:
		return fmt.Errorf("unknown file extension %q", ext)
	}

	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	rows := make(chan interface{})
	done := make(chan error)

	go func() {
		done <- f.dumper.Dump(f.Type, f.language, rows)
	}()

	go func() {
		for scanner.Scan() {
			line := strings.TrimSuffix(scanner.Text(), ",")
			if line == "[" || line == "]" {
				continue
			}

			l := []byte(line)

			// skip every other line e.g. {"index":{"_type":"page","_id":"17949905"}}
			tmp := &struct {
				Index struct {
					ID string `json:"_id"`
					T  string `json:"_type"`
				} `json:"index"`
			}{}

			if err := json.Unmarshal(l, tmp); err != nil {
				done <- err
			}

			if tmp.Index.ID != "" {
				continue
			}

			var w interface{}

			switch f.Type {
			case WikidataFT:
				w = &Wikidata{}
				if err := json.Unmarshal(l, w); err != nil {
					done <- err
				}
			case WikipediaFT:
				// Note: there are some duplicates ID's in the files.
				// Also some don't have a wikibase_item (w.ID="").
				w = &Wikipedia{truncate: truncate}
				if err := json.Unmarshal(l, w); err != nil {
					done <- err
				}
			case WikiquoteFT:
				w = &Wikiquote{}
				if err := json.Unmarshal(l, w); err != nil {
					done <- err
				}
			case WiktionaryFT:
				w = &Wiktionary{}
				if err := json.Unmarshal(l, w); err != nil {
					done <- err
				}
			}

			rows <- w
		}

		if err := scanner.Err(); err != nil {
			done <- err
		}

		close(rows)
	}()

	return <-done
}

var reWikipedia = regexp.MustCompile(`^([a-z_]+)wiki-\d{8}-cirrussearch-content.json.gz$`)
var reWikiquote = regexp.MustCompile(`^([a-z_]+)wikiquote-\d{8}-cirrussearch-content.json.gz$`)
var reWiktionary = regexp.MustCompile(`^([a-z_]+)wiktionary-\d{8}-cirrussearch-content.json.gz$`)

// CirrusLinks finds the latest cirrus links available from wikipedia.
// e.g. enwiki-20171009-cirrussearch-content.json.gz
// Note: Cirrus is their elasticsearch-formatted dump files. The cirrussearch urls
// for wikipedia includes the wikibase_item and has a more similar
// layout to their API than the dumps found at https://dumps.wikimedia.org/enwiki/latest/.
func CirrusLinks(supported []language.Tag, fileTypes []FileType) ([]*File, error) {
	resp, err := http.Get(CirrusURL.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)
	var tt html.TokenType

	var files = []*File{}

	for {
		tt = z.Next()

		switch tt {
		case html.ErrorToken:
			return files, nil
		case html.StartTagToken, html.SelfClosingTagToken:
			t := z.Token()
			if t.DataAtom == atom.A {
				for _, a := range t.Attr {
					if a.Key == "href" {
						for _, ft := range fileTypes {
							var re *regexp.Regexp

							switch ft {
							case WikipediaFT:
								re = reWikipedia
							case WikiquoteFT:
								re = reWikiquote
							case WiktionaryFT:
								re = reWiktionary
							default:
								return nil, fmt.Errorf("unknown filetype %q", ft)
							}

							match := re.FindStringSubmatch(a.Val)

							if len(match) != 2 { // e.g. [enwiki-20171023-cirrussearch-content.json.gz en]
								continue
							}

							if lang, ok := isSupported(match[1], supported); ok {
								u, err := url.Parse(a.Val)
								if err != nil {
									return nil, err
								}

								u = CirrusURL.ResolveReference(u)

								f := NewFile(u, ft, lang)
								files = append(files, f)
							}
						}
					}
				}
			}
		}
	}
}

// supported checks to see if the language of the file is supported
func isSupported(w string, supported []language.Tag) (language.Tag, bool) {
	// check to see if the language is supported
	var valid bool

	// skip some files that we will never need to download
	skip := []string{
		"advisory", "be_x_old", "commons", "donate", "fixcopyright", "foundation", "incubator",
		"labs", "labtest", "login", "mediawiki", "meta", "nostalgia", "outreach",
		"quality", "species", "simple", "sources", "strategy", "test",
		"testwikidata", "usability", "vote", "wikidata", // we get wikidata from another source ;)
		"atj", "eml", "roa_tara", "ten", "zh_classical", "wikimania",
		"testcommons", "hyw",
	}

	for _, s := range skip {
		if w == s {
			return language.Tag{}, valid
		}
	}

	lang := language.MustParse(w)
	for _, l := range supported {
		if l == lang {
			valid = true
			break
		}
	}

	return lang, valid
}
