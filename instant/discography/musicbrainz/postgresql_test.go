package musicbrainz

import (
	"database/sql/driver"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/jivesearch/jivesearch/instant/discography"
	"github.com/lib/pq"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v2"
)

func TestFetch(t *testing.T) {
	type args struct {
		artist  string
		artmbid string
		row1    []driver.Value
		row2    []driver.Value
	}

	u, _ := url.Parse("http://coverartarchive.org/release/art_mbid/1-250..jpg")

	tests := []struct {
		name string
		args args
		want []discography.Album
	}{
		{
			"basic",
			args{
				"matisyahu",
				"art_mbid",
				[]driver.Value{"1", "Album 1", "art_mbid", time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC)},
				[]driver.Value{"art_mbid", "1", ".jpg"},
			},
			[]discography.Album{
				{
					Name:      "Album 1",
					Published: time.Date(1998, 1, 1, 0, 0, 0, 0, time.UTC),
					Image: discography.Image{
						URL: u,
					},
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
				[]string{"gid", "name", "art_mbid", "published"},
			)

			rows = rows.AddRow(
				tt.args.row1...,
			)

			mock.ExpectQuery("SELECT").WithArgs(tt.args.artist).WillReturnRows(rows)

			rows = sqlmock.NewRows(
				[]string{"gid", "id", "suffix"},
			)
			rows = rows.AddRow(
				tt.args.row2...,
			)

			mock.ExpectQuery("SELECT").WithArgs(pq.Array([]string{tt.args.artmbid})).WillReturnRows(rows)

			p := &PostgreSQL{
				DB: db,
			}

			got, err := p.Fetch(tt.args.artist)
			if err != nil {
				t.Fatal(err)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}
