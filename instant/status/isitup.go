package status

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// IsItUp holds settings for the isitup.org API
type IsItUp struct {
	HTTPClient *http.Client
}

// IsItUpProvider indicates the source is isitup.org
const IsItUpProvider provider = "Is it up?"

// Fetch retrieves security breaches from haveibeenpwned.com
func (u *IsItUp) Fetch(domain string) (*Response, error) {
	uu, err := url.Parse(fmt.Sprintf("https://isitup.org/%v.json", domain))
	if err != nil {
		return nil, err
	}

	resp, err := u.HTTPClient.Get(uu.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	r := &Response{
		Provider: IsItUpProvider,
	}

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err

}
