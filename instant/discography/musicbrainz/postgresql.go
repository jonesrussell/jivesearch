package musicbrainz

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/jonesrussell/jivesearch/instant/discography"
	"github.com/lib/pq"
)

// PostgreSQL holds configuration for a MusicBrainz database
type PostgreSQL struct {
	*sql.DB
}

// Fetch fetches MusicBrainz albums and coverart from Postgresql
func (p *PostgreSQL) Fetch(artist string) ([]discography.Album, error) {
	var albums = []discography.Album{}
	var err error

	// The following combines some of the sql found at:
	// https://github.com/metabrainz/musicbrainz-server/blob/master/lib/MusicBrainz/Server/Data/ReleaseGroup.pm
	// https://github.com/metabrainz/coverart_redirect/blob/master/coverart_redirect/request.py
	sql := `
		WITH albums AS (
			SELECT DISTINCT rg.gid, rg.name, 
				COALESCE(art.mbid::text,'') art_mbid,
				make_date(
					coalesce(rgm.first_release_date_year, 0001), 
					coalesce(rgm.first_release_date_month, 1),
					coalesce(rgm.first_release_date_day, 1)
				) published, 
				array(
					SELECT COALESCE(id, 0) FROM musicbrainz.release_group_secondary_type rgst
					JOIN musicbrainz.release_group_secondary_type_join rgstj
					ON rgstj.secondary_type = rgst.id
					WHERE rgstj.release_group = rg.id
					ORDER BY name ASC
				) secondary_types				
			FROM musicbrainz.release_group rg
			JOIN musicbrainz.release_group_meta rgm ON rgm.id = rg.id
			LEFT JOIN LATERAL (
				SELECT DISTINCT ON (release.release_group) release.gid AS mbid, release_group.gid
				FROM musicbrainz.index_listing
				JOIN musicbrainz.release ON musicbrainz.release.id = musicbrainz.index_listing.release
				JOIN musicbrainz.release_group ON release_group.id = release.release_group
				LEFT JOIN (
					SELECT release, date_year, date_month, date_day
					FROM musicbrainz.release_country
					UNION ALL
					SELECT release, date_year, date_month, date_day
					FROM musicbrainz.release_unknown_country
				) release_event ON (release_event.release = release.id)
				FULL OUTER JOIN musicbrainz.release_group_cover_art ON release_group_cover_art.release = musicbrainz.release.id
				WHERE release_group.gid = rg.gid
				AND is_front = true
				ORDER BY release.release_group, release_group_cover_art.release,
					release_event.date_year, release_event.date_month, release_event.date_day
			) art ON true
			WHERE rg.artist_credit=(
				SELECT id FROM musicbrainz.artist a 
				WHERE a.gid=(
					SELECT 
					(wd."claims"->'musicbrainz'->>0)::uuid mbid
					FROM enwikipedia w
					LEFT JOIN wikidata wd ON w.id = wd.id			
					WHERE LOWER(w.title) = LOWER($1)
					AND jsonb_array_length(claims->'musicbrainz') > 0
					LIMIT 1
				)
			)
			AND rg.type=1
			ORDER BY published
		)
		
		SELECT gid, name, art_mbid, published
		FROM albums
		WHERE albums.secondary_types = '{}'
	`

	rows, err := p.DB.Query(sql, artist)
	if err != nil {
		return albums, err
	}

	var artMBIDs = map[string]string{}

	defer rows.Close()
	for rows.Next() {
		var artMBID string

		a := discography.Album{}
		if err := rows.Scan(&a.ID, &a.Name, &artMBID, &a.Published); err != nil {
			return albums, err
		}

		albums = append(albums, a)
		artMBIDs[artMBID] = a.ID
	}

	// We could put the following in the above statement using a lateral join but it is slow (or I didn't do it right).
	sql = `
		SELECT musicbrainz.release.gid, index_listing.id, image_type.suffix
		FROM musicbrainz.index_listing
		JOIN musicbrainz.release
		ON musicbrainz.index_listing.release = musicbrainz.release.id
		JOIN musicbrainz.image_type
		ON musicbrainz.index_listing.mime_type = musicbrainz.image_type.mime_type
		WHERE musicbrainz.release.gid::text = ANY ($1)
		AND is_front = true
		ORDER BY ordering
	`

	var m = []string{}
	for k := range artMBIDs {
		m = append(m, k)
	}

	rows, err = p.DB.Query(sql, pq.Array(m))
	if err != nil {
		return albums, err
	}

	defer rows.Close()
	for rows.Next() {
		var mbid string
		var id string
		var ext string

		if err := rows.Scan(&mbid, &id, &ext); err != nil {
			return albums, err
		}

		u, _ := url.Parse(fmt.Sprintf("http://coverartarchive.org/release/%v/%v-250.%v", mbid, id, ext))
		img := discography.Image{
			URL: u,
		}

		for i, a := range albums {
			if a.ID != artMBIDs[mbid] {
				continue
			}

			albums[i].Image = img
		}
	}

	return albums, err
}
