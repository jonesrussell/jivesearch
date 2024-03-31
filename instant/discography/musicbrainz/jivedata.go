package musicbrainz

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/jonesrussell/jivesearch/instant/discography"
)

// JiveData is a Wikipedia data provider
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// Fetch retrieves Wikipedia Items from Jive Data
func (j *JiveData) Fetch(artist string) ([]discography.Album, error) {
	u, err := url.Parse("https://jivedata.com/musicbrainz")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("artist", artist)
	u.RawQuery = q.Encode()

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	d := []discography.Album{}
	err = json.NewDecoder(resp.Body).Decode(&d)

	return d, err
}
