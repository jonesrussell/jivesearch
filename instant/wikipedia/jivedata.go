package wikipedia

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/text/language"
)

// JiveData is a Wikipedia data provider
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// Fetch retrieves Wikipedia Items from Jive Data
func (j *JiveData) Fetch(query string, lang language.Tag) ([]*Item, error) {
	u, err := url.Parse("https://jivedata.com/wikipedia")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("q", query)
	q.Set("l", lang.String())
	u.RawQuery = q.Encode()

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	w := []*Item{}
	err = json.NewDecoder(resp.Body).Decode(&w)

	return w, err
}

// Setup performs setup actions
func (j *JiveData) Setup() error {
	return nil
}
