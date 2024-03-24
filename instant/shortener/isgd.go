package shortener

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/jivesearch/jivesearch/log"
)

// IsGd holds settings to shorten urls from is.gd
type IsGd struct {
	HTTPClient *http.Client
}

// IsGdProvider is a url shortening service
var IsGdProvider provider = "is.gd"

// Shorten shortens a url
func (g *IsGd) Shorten(u *url.URL) (*Response, error) {
	uu := fmt.Sprintf("https://is.gd/create.php?format=json&url=%v", u.String())
	fmt.Println(uu)

	resp, err := g.HTTPClient.Get(uu)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := &Response{
		Original: u,
		Provider: IsGdProvider,
	}

	type tmp struct {
		ShortURL  string `json:"shorturl"`
		ErrorCode int    `json:"errorcode"`
	}

	t := &tmp{}

	if err := json.Unmarshal(bdy, &t); err != nil {
		log.Info.Println(err)
	}

	if t.ErrorCode != 0 {
		return nil, fmt.Errorf("unable to shorten url")
	}

	r.Short, err = url.Parse(t.ShortURL)
	return r, err
}
