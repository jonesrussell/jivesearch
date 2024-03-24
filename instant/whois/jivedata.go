package whois

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// JiveData retrieves WHOIS information from Jive Data
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// JiveDataProvider is a WHOIS provider
var JiveDataProvider provider = "Jive Data"

// Fetch retrieves from the IEX api
func (j *JiveData) Fetch(domain string) (*Response, error) {
	u, err := url.Parse("https://jivedata.com/whois")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("domain", domain)
	u.RawQuery = q.Encode()

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	r := &Response{}
	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}
