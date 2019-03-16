package wikipedia

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jivesearch/jivesearch/log"
	"github.com/lib/pq"
	"golang.org/x/text/language"
)

// PostgreSQL contains our client and database info
type PostgreSQL struct {
	*sql.DB
}

type tableType = string

const wikidataAliasesTable tableType = "wikidata_aliases"
const wikidataTable tableType = "wikidata"
const wikipediaTable tableType = "wikipedia"
const wikiquoteTable tableType = "wikiquote"
const wiktionaryTable tableType = "wiktionary"

type table struct {
	Type      tableType
	name      string
	temporary string
	columns   []column
	rows      chan interface{}
}

type column struct {
	name  string
	t     string
	index bool
}

// Scan unmarshals jsonb data
func (l *Labels) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), l)
}

// Scan unmarshals jsonb data
func (a *Aliases) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), a)
}

// Scan unmarshals jsonb data
func (d *Descriptions) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), d)
}

// Scan unmarshals jsonb data
// http://www.booneputney.com/development/gorm-golang-jsonb-value-copy/
func (c *Claims) Scan(value interface{}) error {
	return json.Unmarshal(value.([]byte), c)
}

// Fetch retrieves an Item from PostgreSQL
// https://www.wikidata.org/w/api.php
func (p *PostgreSQL) Fetch(query string, lang language.Tag) ([]*Item, error) {
	item := &Item{
		Wikipedia: Wikipedia{
			Language: lang.String(),
		},
		Wikidata: &Wikidata{
			Claims: &Claims{},
		},
		Wiktionary: Wiktionary{
			Language: lang.String(),
		},
	}

	// iterate through Data
	tags, objects, stmts := []string{}, []string{}, []string{}

	data := reflect.Indirect(reflect.ValueOf(item.Claims))
	for i := 0; i < data.NumField(); i++ {
		tag := strings.Split(data.Type().Field(i).Tag.Get("json"), ",")[0]

		tags = append(tags, fmt.Sprintf(`"%v"`, tag))
		objects = append(objects, fmt.Sprintf(`jsonb_build_object('%v', "%v".item)`, tag, tag)) // `'influences', influences."item"`

		switch data.Field(i).Interface().(type) {
		case []string, []Text:
			switch tag {
			case "image", "flag":
				/* two urls to get images:
				1. https://commons.wikimedia.org/w/thumb.php?width=500&f=Junior-Jaguar-Belize-Zoo.jpg
					a. Will resize the image. Requesting a larger size than the original will result in an error
				2. https://upload.wikimedia.org/wikipedia/commons/2/21/Junior-Jaguar-Belize-Zoo.jpg
					a. 2 & 21 represent the first and first two characters of the image md5 hash
				*/
				stmts = append(stmts, fmt.Sprintf(`
					"%v" AS (
						SELECT build_image(claims->'%v') item
						FROM item
					)`, tag, tag),
				)
			default:
				stmts = append(stmts, fmt.Sprintf(`
					"%v" AS (
						SELECT claims->'%v' item
						FROM item
					)`, tag, tag),
				)
			}

		case []DateTime:
			stmts = append(stmts, fmt.Sprintf(`
				"%v" AS (
					SELECT build_datetime(claims->'%v') item
					FROM item
				)`, tag, tag),
			)
		case []Quantity:
			stmts = append(stmts, fmt.Sprintf(`
				"%v" AS (
					SELECT build_quantity(claims->'%v') item
					FROM item
				)`, tag, tag),
			)
		case []Wikidata:
			stmts = append(stmts, fmt.Sprintf(`
				"%v" AS (
					SELECT build_item(claims->'%v') item
					FROM item
				)`, tag, tag),
			)
		case []Coordinate:
			stmts = append(stmts, fmt.Sprintf(`
				"%v" AS (
					SELECT build_coordinate(claims->'%v') item
					FROM item
				)`, tag, tag),
			)
		default: // e.g. has qualifiers
			var elem reflect.Value
			field := reflect.Indirect(reflect.ValueOf(item.Claims)).Field(i)

			typ := field.Type().Elem()
			if typ.Kind() == reflect.Ptr {
				elem = reflect.New(typ.Elem())
			}
			if typ.Kind() == reflect.Struct {
				elem = reflect.New(typ).Elem()
			}

			var inner []string

			for j := 0; j < reflect.Indirect(elem).NumField(); j++ {
				t := strings.Split(elem.Type().Field(j).Tag.Get("json"), ",")[0]

				switch elem.Field(j).Interface().(type) {
				case []string:
					inner = append(inner, fmt.Sprintf("'%v', x.d->'%v'", t, t))
				case []Wikidata:
					inner = append(inner, fmt.Sprintf("'%v', build_item(x.d->'%v')", t, t))
				case []Quantity:
					inner = append(inner, fmt.Sprintf("'%v', build_quantity(x.d->'%v')", t, t))
				case []DateTime:
					inner = append(inner, fmt.Sprintf("'%v', build_datetime(x.d->'%v')", t, t))
				default:
					log.Info.Printf(" unsupported field for %v\n", t)

				}
			}

			stmts = append(stmts, fmt.Sprintf(`
				"%v" AS (
					SELECT jsonb_agg(
						jsonb_build_object(
							%v
						)
					) item
					FROM item
					JOIN LATERAL (
						SELECT * FROM jsonb_array_elements(item.claims->'%v')
					) as x(d) on true
				)`, tag, strings.Join(inner, ", "), tag,
			))
		}
	}

	// Note: We cannot build 1 large jsonb_build_object as PostgreSQL has a 100 item limit.
	sql := fmt.Sprintf(`
		WITH item AS (
			SELECT *
			FROM (
				SELECT 
				w."id", w."title", w."text", w."outgoing_link", wq."quotes", 
				wd."labels", wd."descriptions", wd."claims" 
				FROM %vwikipedia w
				LEFT JOIN %vwikiquote wq ON w.id = wq.id
				LEFT JOIN wikidata wd ON w.id = wd.id			
				WHERE LOWER(w.title) = LOWER($1)
				/* A UNION seems to be faster than OR here */
				UNION
				SELECT
				w."id", w."title", w."text", w."outgoing_link", wq."quotes", 
				wd."labels", wd."descriptions", wd."claims" 
				FROM enwikipedia w
				LEFT JOIN enwikiquote wq ON w.id = wq.id
				LEFT JOIN wikidata wd ON w.id = wd.id
				LEFT JOIN wikidata_aliases wa ON w.id = wa.id		
				WHERE LOWER(wa.alias) = LOWER($2)
				AND wa.lang = '%v'
				LIMIT 1
			) w
			FULL OUTER JOIN (
				SELECT "title" wktitle, "definitions"
				FROM %vwiktionary
				WHERE title = $3
				LIMIT 1
			) wk ON LOWER(w.title) = wk.wktitle
		),
		%v
		SELECT
			coalesce(item."id", ''), coalesce(item."title", ''), coalesce(item."text", ''), coalesce(item."outgoing_link", '{}'),
			coalesce(item."quotes", '{}'), coalesce(item."wktitle", ''), coalesce(item."definitions", '[]'),
			coalesce(item."labels", '{}'::jsonb), coalesce(item."descriptions", '{}'::jsonb), %v "claims"
		FROM item, %v
	`, item.Wikipedia.Language, item.Wiktionary.Language, item.Wikipedia.Language, item.Wiktionary.Language,
		strings.Join(stmts, ", "), strings.Join(objects, " || "), strings.Join(tags, ", "),
	)

	var definitions string

	err := p.DB.QueryRow(sql, query, query, query).Scan(
		&item.Wikidata.ID, &item.Wikipedia.Title, &item.Wikipedia.Text, pq.Array(&item.Wikipedia.OutgoingLink),
		pq.Array(&item.Wikiquote.Quotes), &item.Wiktionary.Title, &definitions,
		&item.Labels, &item.Descriptions, &item.Claims,
	)

	if err != nil {
		return []*Item{item}, err
	}

	//log.Debug.Println(sql)

	err = json.Unmarshal([]byte(definitions), &item.Wiktionary.Definitions)

	// is it a disambiguation page???
	if v, ok := item.Wikidata.Descriptions["en"]; ok {
		if v.Text == "Wikipedia disambiguation page" || v.Text == "Wikimedia disambiguation page" {
			dis := []string{}
			lc := strings.ToLower(strings.Replace(item.Wikipedia.Title, " ", "_", -1))

			for _, d := range item.OutgoingLink {
				if strings.HasPrefix(strings.ToLower(d), lc+"_") || strings.HasPrefix(strings.ToLower(d), lc+",_") { // e.g. Sublime,_Texas w/ a comma
					d = strings.Replace(d, "_", " ", -1)
					dis = append(dis, strings.ToLower(d))
				}
			}

			sql = fmt.Sprintf(`SELECT w.id, title, text, popularity_score
				FROM %vwikipedia w
				LEFT JOIN wikidata wd ON w.ID=wd.ID
				WHERE LOWER(w.title) = ANY($1)
				ORDER BY popularity_score DESC
				LIMIT 10
			`, item.Wikipedia.Language)

			items := []*Item{}

			rows, err := p.DB.Query(sql, pq.Array(dis))
			if err != nil {
				return []*Item{item}, err
			}

			//log.Debug.Println(pq.Array(dis))

			defer rows.Close()
			for rows.Next() {
				item := &Item{
					Wikipedia: Wikipedia{
						Language: lang.String(),
					},
					Wikidata: &Wikidata{
						Claims: &Claims{},
					},
					Wiktionary: Wiktionary{
						Language: lang.String(),
					},
				}
				var pop float64
				if err := rows.Scan(&item.Wikidata.ID, &item.Wikipedia.Title, &item.Wikipedia.Text, &pop); err != nil {
					return []*Item{item}, err
				}

				items = append(items, item)
			}
			if err := rows.Err(); err != nil {
				return items, err
			}

			return items, err
		}
	}

	return []*Item{item}, err
}

type transaction = func(tx *sql.Tx) error

func (p *PostgreSQL) executeTransaction(t transaction) (err error) {
	tx, err := p.DB.Begin()
	if err != nil {
		panic(err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}

		if e := tx.Commit(); e != nil {
			err = e
		}
	}()

	err = t(tx)
	return
}

// Dump creates a temporary table and dumps rows via our transaction
func (p *PostgreSQL) Dump(ft FileType, lang language.Tag, rows chan interface{}) error {
	t := &table{
		rows: rows,
	}

	var aliases *table

	switch ft {
	case WikidataFT:
		t.Type = wikidataTable
		t.name = wikidataTable
		aliases = &table{
			Type:      wikidataAliasesTable,
			name:      wikidataAliasesTable,
			temporary: wikidataAliasesTable + "_tmp",
		}
		aliases.setColumns()
	case WikipediaFT:
		t.Type = wikipediaTable
		n := strings.Replace(lang.String(), "-", "_", -1)
		t.name = fmt.Sprintf("%v%v", strings.ToLower(n), wikipediaTable) // enwikipedia, cebwikipedia, etc...
	case WikiquoteFT:
		t.Type = wikiquoteTable
		n := strings.Replace(lang.String(), "-", "_", -1)
		t.name = fmt.Sprintf("%v%v", strings.ToLower(n), wikiquoteTable) // enwikiquote, cebwikiquote, etc...
	case WiktionaryFT:
		t.Type = wiktionaryTable
		n := strings.Replace(lang.String(), "-", "_", -1)
		t.name = fmt.Sprintf("%v%v", strings.ToLower(n), wiktionaryTable) // enwiktionary, cebwiktionary, etc...
	default:
		return fmt.Errorf("unknown filetype %q", ft)
	}

	t.temporary = t.name + "_tmp"

	if err := t.setColumns(); err != nil {
		return err
	}

	if _, err := p.DB.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %v`, t.temporary)); err != nil {
		return err
	}

	if _, err := p.DB.Exec(t.createTable()); err != nil {
		return err
	}

	if err := p.executeTransaction(t.insert); err != nil {
		return err
	}

	if err := p.executeTransaction(t.addIndices); err != nil {
		return err
	}

	if err := p.executeTransaction(t.rename); err != nil {
		return err
	}

	if aliases != nil {
		if _, err := p.DB.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %v`, aliases.temporary)); err != nil {
			return err
		}

		if _, err := p.DB.Exec(aliases.createTable()); err != nil {
			return err
		}

		stmt := fmt.Sprintf(`INSERT INTO %v(id, lang, alias)
		(
			SELECT id, b.el->>'language', b.el->>'value'
			FROM
			(
				SELECT id, jsonb_array_elements(v) AS el
				FROM
				(
					SELECT id, t.k, t.v
					FROM %v, jsonb_each(aliases) as t(k, v)
				) a
			) b
		 )		 
		 `, aliases.temporary, t.name)
		if _, err := p.DB.Exec(stmt); err != nil {
			return err
		}

		if err := p.executeTransaction(aliases.addIndices); err != nil {
			return err
		}

		if err := p.executeTransaction(aliases.rename); err != nil {
			return err
		}

		if _, err := p.DB.Exec(fmt.Sprintf(`ALTER TABLE %v DROP COLUMN aliases`, t.name)); err != nil {
			return err
		}
	}

	return nil
}

func (t *table) setColumns() error {
	var err error

	switch t.Type {
	case wikidataAliasesTable:
		t.columns = []column{
			{"id", "text", true},
			{"lang", "text", true},
			{"alias", "text", true},
		}
	case wikidataTable:
		t.columns = []column{
			{"id", "text", true},
			{"labels", "jsonb", false},
			{"aliases", "jsonb", false},
			{"descriptions", "jsonb", false},
			{"claims", "jsonb", false},
		}
	case wikipediaTable:
		t.columns = []column{
			{"id", "text", true},
			{"title", "text", true},
			{"text", "text", false},
			{"outgoing_link", "text[]", false},
			{"popularity_score", "numeric", true},
		}
	case wikiquoteTable:
		t.columns = []column{
			{"id", "text", true},
			{"quotes", "text[]", false},
		}
	case wiktionaryTable:
		t.columns = []column{
			{"title", "text", true},
			{"definitions", "jsonb", false},
		}
	default:
		err = fmt.Errorf("unknown postgresql table type %v", t.Type)
	}

	return err
}

func (t *table) createTable() string {
	c := fmt.Sprintf(`CREATE TABLE %v (pk serial PRIMARY KEY,`, t.temporary)

	cols := []string{}
	for _, col := range t.columns {
		switch col.name {
		case "outgoing_link":
			cols = append(cols, fmt.Sprintf("%v %v", col.name, col.t))
		default:
			cols = append(cols, fmt.Sprintf("%v %v NOT NULL", col.name, col.t))
		}
	}

	c += strings.Join(cols, ", ") + ")"
	return c
}

func (t *table) insert(tx *sql.Tx) (err error) {
	cols := []string{}
	for _, col := range t.columns {
		cols = append(cols, col.name)
	}

	stmt, err := tx.Prepare(pq.CopyIn(t.temporary, cols...))
	if err != nil {
		return
	}

	defer func() {
		if e := stmt.Close(); err == nil && e != nil {
			err = e
		}
	}()

	// dump the rows
	for row := range t.rows {
		r := []interface{}{}

		switch row := row.(type) {
		case *Wikipedia:
			r = []interface{}{row.ID, row.Title, row.Text, pq.Array(row.OutgoingLink), row.Popularity}
		case *Wikidata:
			r = []interface{}{row.ID}

			// The following are all jsonb columns.
			val := reflect.ValueOf(row).Elem()
			for i := 1; i < val.NumField(); i++ {
				j, e := json.Marshal(val.Field(i).Interface())
				if e != nil {
					err = e
					return
				}

				r = append(r, string(j))
			}
		case *Wikiquote:
			if len(row.Quotes) == 0 {
				continue
			}
			r = []interface{}{row.ID, pq.Array(row.Quotes)}
		case *Wiktionary:
			r = []interface{}{row.Title}

			// jsonb column
			j, e := json.Marshal(row.Definitions)
			if e != nil {
				err = e
				return
			}
			r = append(r, string(j))

		default:
			err = fmt.Errorf("unknown datatype for %+v", r)
			return
		}

		if _, err = stmt.Exec(r...); err != nil {
			return
		}
	}

	return
}

func (t *table) indexName(tbl, col string) string {
	return fmt.Sprintf("%v_%v", tbl, col)
}

// addIndices adds indexes to our temporary table
func (t *table) addIndices(tx *sql.Tx) (err error) {
	for _, c := range t.columns {
		if !c.index {
			continue
		}

		col := c.name
		if (t.Type == wikipediaTable && c.name == "title") || (t.Type == wikidataAliasesTable && c.name == "alias") {
			col = fmt.Sprintf("LOWER(%v)", col)
		}

		var using string
		if c.t == "jsonb" {
			using = "USING gin"
		}

		idx := fmt.Sprintf(`CREATE INDEX %v ON %v %v (%v)`, t.indexName(t.temporary, c.name), t.temporary, using, col)
		if _, err = tx.Exec(idx); err != nil {
			return err
		}
	}

	return err
}

// rename renames the t.temporary table to t.name
func (t *table) rename(tx *sql.Tx) (err error) {
	_, err = tx.Exec(fmt.Sprintf(`DROP TABLE IF EXISTS %v`, t.name))
	if err != nil {
		return err
	}

	_, err = tx.Exec(fmt.Sprintf(`ALTER TABLE %v RENAME to %v`, t.temporary, t.name))
	if err != nil {
		return err
	}

	for _, c := range t.columns {
		if !c.index {
			continue
		}

		_, err = tx.Exec(fmt.Sprintf(`ALTER INDEX %v RENAME to %v`,
			t.indexName(t.temporary, c.name), t.indexName(t.name, c.name)),
		)
		if err != nil {
			return err
		}
	}

	return err
}

// Setup creates our functions
func (p *PostgreSQL) Setup() error {
	buildItem := `
	CREATE OR REPLACE FUNCTION build_item(jsonb) 
	RETURNS jsonb immutable strict language sql as $$
	   SELECT jsonb_agg(                                               
			jsonb_build_object(                                     
				 'id', x.d->'id',                                
				 'labels', wikidata.labels,
				 'claims', coalesce(wikidata.claims, '{}'::jsonb)              
			)                                                       
		  )                                                               
		FROM jsonb_array_elements($1) AS x(d)                           
	   LEFT JOIN wikidata ON (x.d->>'id') = wikidata.id     
	$$;
	`

	buildCoordinate := `
	CREATE OR REPLACE FUNCTION build_coordinate(jsonb) 
	RETURNS jsonb immutable strict language sql as $$
		SELECT jsonb_agg(                                               
			jsonb_build_object(                                     
				'latitude', x.d->'latitude',
				'longitude', x.d->'longitude',
				'altitude', x.d->'altitude',
				'precision', x.d->'precision',				
				'globe', x.d->'globe'                
			)                                                       
		)                                                               
		FROM jsonb_array_elements($1) AS x(d)                           
	$$; 
	`

	buildDateTime := `
	CREATE OR REPLACE FUNCTION build_datetime(jsonb) 
	RETURNS jsonb immutable strict language sql as $$
		SELECT jsonb_agg(                                               
			jsonb_build_object(                                     
				'value', x.d->'value',                          
				'calendar', jsonb_build_object(                 
					'id', x.d->'calendar'->>'id',           
					'labels', wikidata.labels               
				)                                               
			)                                                       
		)                                                               
		FROM jsonb_array_elements($1) AS x(d)                           
		LEFT JOIN wikidata on (x.d->'calendar'->>'id') = wikidata.id 
	$$;
	`

	buildQuantity := `
	CREATE OR REPLACE FUNCTION build_quantity(jsonb) 
	RETURNS jsonb immutable strict language sql as $$
		SELECT jsonb_agg(                                               
			jsonb_build_object(                                     
				'amount', x.d->'amount',                        
				'unit', jsonb_build_object(                     
					'id', x.d->'unit'->>'id',               
					'labels', wikidata.labels               
				)                                               
			)                                                       
		)                                                               
		FROM jsonb_array_elements($1) AS x(d)                           
		LEFT JOIN wikidata on (x.d->'unit'->>'id') = wikidata.id       
	$$; 
	`

	/*
		NOTE: using 2 slashes as below e.g. 'https://upload…'
		will result in a 301 redirect to 'https:/upload….'
		then a 200 response….trying to cut out the redirect
		by using just 1 slash  will result in an invalid signature...
		we have to have the redirect for some reason.
	*/
	buildImage := `
	CREATE OR REPLACE FUNCTION build_image(jsonb) 
	RETURNS jsonb immutable strict language sql as $$
	SELECT jsonb_agg(  
		'https://upload.wikimedia.org/wikipedia/commons/' || LEFT(item.m, 1) || '/' || LEFT(item.m, 2) || '/' || s                                  
	)                                                               
	FROM (
		SELECT 
			md5(REPLACE(x.d::text, ' ', '_')) AS m,
			REPLACE(x.d::text, ' ', '_') AS s
		FROM jsonb_array_elements_text($1) AS x(d) 
	) item  
	$$;
	`

	for _, f := range []string{buildItem, buildCoordinate, buildDateTime, buildQuantity, buildImage} {
		if _, err := p.DB.Exec(f); err != nil {
			return err
		}
	}

	return nil
}
