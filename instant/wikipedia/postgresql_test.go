package wikipedia

import (
	"database/sql/driver"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/lib/pq"
	"golang.org/x/text/language"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"
)

var shaqClaimsJSON = []byte(`{
	"sex": [{"id": "Q6581097", "labels": {"en": {"value": "male", "language": "en"}}}]
}`)

var shaqClaimsPostgres = &Claims{
	Sex: []Wikidata{
		{
			ID: "Q6581097",
			Labels: map[string]Text{
				"en": {Text: "male", Language: "en"},
			},
			Claims: &Claims{},
		},
	},
}

// we're just gonna pretend there is a wiktionary entry for Shaquille O'Neal
var shaqWiktionaryJSON = []byte(
	`[{"part": "noun", "meaning": "friendly person", "synonyms": [{"word": "superman", "language": "en"}]}]`,
)

var shaqWiktionary = &Wiktionary{
	Title:    "Shaquille O'Neal",
	Language: "en",
	Definitions: []*Definition{
		{
			Part:    "noun",
			Meaning: "friendly person",
			Synonyms: []Synonym{
				{Language: "en", Word: "superman"},
			},
		},
	},
}

func TestPostgreSQL_Fetch(t *testing.T) {
	type args struct {
		query string
		lang  language.Tag
		row   []driver.Value
	}

	tests := []struct {
		name string
		args args
		want []*Item
	}{
		{
			"shaq",
			args{
				"Shaquille O'Neal", language.MustParse("en"),
				[]driver.Value{
					"Q169452", "Shaquille O'Neal", "Shaquille O'Neal is a basketball player", "{}",
					"{}", "Shaquille O'Neal", shaqWiktionaryJSON,
					[]byte(shaqRawLabels), []byte(shaqRawDescriptions), shaqClaimsJSON,
				},
			},
			[]*Item{
				{
					Wikipedia: Wikipedia{
						Language:     "en",
						OutgoingLink: []string{},
						Title:        "Shaquille O'Neal",
						Text:         "Shaquille O'Neal is a basketball player",
					},
					Wikidata: &Wikidata{
						ID:           "Q169452",
						Descriptions: shaqDescriptions,
						Labels:       shaqLabels,
						Claims:       shaqClaimsPostgres,
					},
					Wikiquote: Wikiquote{
						//Quotes: shaqQuotes, // not really sure how to test for pq.Array
						Quotes: []string{},
					},
					Wiktionary: *shaqWiktionary,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			rows := sqlmock.NewRows(
				[]string{"id", "title", "text", "outgoing_link", "quotes", "wktitle", "definitions",
					"labels", "descriptions", "claims",
				},
			)
			rows = rows.AddRow(
				tt.args.row...,
			)

			mock.ExpectQuery("SELECT").WithArgs(tt.args.query, tt.args.query, tt.args.query).WillReturnRows(rows)

			p := &PostgreSQL{
				DB: db,
			}

			got, err := p.Fetch(tt.args.query, tt.args.lang)
			if err != nil {
				t.Fatal(err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got[0], tt.want[0]) {
				t.Errorf("got %+v, want %+v", got[0], tt.want[0])
			}
		})
	}
}

func TestPostgreSQL_Dump(t *testing.T) {
	type args struct {
		lang language.Tag
		ft   FileType
	}
	tests := []struct {
		name string
		row  interface{}
		args args
	}{
		{
			"enwiki",
			shaqWikipedia,
			args{
				ft:   WikipediaFT,
				lang: language.MustParse("en"),
			},
		},
		{
			"wikidata",
			shaqWikidata,
			args{
				ft: WikidataFT,
			},
		},
		{
			"enwikiquote",
			shaqWikiquote,
			args{
				ft:   WikiquoteFT,
				lang: language.MustParse("en"),
			},
		},
		{
			"enwiktionary",
			guitarWiktionary,
			args{
				ft:   WiktionaryFT,
				lang: language.MustParse("en"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var testAliases bool

			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer db.Close()

			// create table
			mock.ExpectExec("DROP TABLE IF EXISTS").WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(1, 1))

			// insert data
			mock.ExpectBegin()
			switch tt.args.ft {
			case WikipediaFT:
				w := tt.row.(*Wikipedia)

				mock.ExpectPrepare("COPY").ExpectExec().
					WithArgs(w.ID, w.Title, w.Text, pq.Array(w.OutgoingLink), w.Popularity).
					WillReturnResult(sqlmock.NewResult(1, 1))
			case WikidataFT:
				wd := tt.row.(*Wikidata)

				l, err := json.Marshal(wd.Labels)
				if err != nil {
					t.Fatal(err)
				}
				a, err := json.Marshal(wd.Aliases)
				if err != nil {
					t.Fatal(err)
				}
				d, err := json.Marshal(wd.Descriptions)
				if err != nil {
					t.Fatal(err)
				}
				c, err := json.Marshal(wd.Claims)
				if err != nil {
					t.Fatal(err)
				}

				mock.ExpectPrepare("COPY").ExpectExec().
					WithArgs("Q169452", string(l), string(a), string(d), string(c)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			case WikiquoteFT:
				wq := tt.row.(*Wikiquote)

				mock.ExpectPrepare("COPY").ExpectExec().
					WithArgs("Q169452", pq.Array(wq.Quotes)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			case WiktionaryFT:
				d, err := json.Marshal(tt.row.(*Wiktionary).Definitions)
				if err != nil {
					t.Fatal(err)
				}

				mock.ExpectPrepare("COPY").ExpectExec().
					WithArgs("guitar", string(d)).
					WillReturnResult(sqlmock.NewResult(1, 1))
			default:
				t.Fatalf("unexpected file type %v\n", tt.args.ft)
			}
			mock.ExpectCommit()

			// create indices
			mock.ExpectBegin()

			switch tt.args.ft {
			case WikipediaFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikidataFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikiquoteFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WiktionaryFT:
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			}

			mock.ExpectCommit()

			// rename table
			mock.ExpectBegin()
			mock.ExpectExec("DROP TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			mock.ExpectExec("ALTER TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			switch tt.args.ft {
			case WikipediaFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WikidataFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				testAliases = true
			case WikiquoteFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			case WiktionaryFT:
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			}
			mock.ExpectCommit()

			if testAliases {
				mock.ExpectExec("DROP TABLE IF EXISTS").WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE TABLE").WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectExec("INSERT INTO").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectBegin()
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("CREATE INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mock.ExpectBegin()
				mock.ExpectExec("DROP TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectExec("ALTER INDEX").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
				mock.ExpectCommit()

				mock.ExpectExec("ALTER TABLE").WithArgs().WillReturnResult(sqlmock.NewResult(1, 1))
			}

			p := &PostgreSQL{
				DB: db,
			}

			rows := make(chan interface{})

			go func() {
				rows <- tt.row
				close(rows)
			}()

			if err := p.Dump(tt.args.ft, tt.args.lang, rows); err != nil {
				t.Fatal(err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestSetup(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	p := &PostgreSQL{
		DB: db,
	}

	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("CREATE OR REPLACE FUNCTION").WillReturnResult(sqlmock.NewResult(1, 1))

	err = p.Setup()
	if err != nil {
		t.Fatal(err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
