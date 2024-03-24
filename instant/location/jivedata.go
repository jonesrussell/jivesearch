package location

import (
	"encoding/json"
	"net"
	"net/http"
	"net/url"
)

// JiveData is a location data provider
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// Fetch gets geolocation data from an IP Address
func (j *JiveData) Fetch(ip net.IP) (*City, error) {
	u, err := url.Parse("https://jivedata.com/geolocation")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("ip", ip.String())
	u.RawQuery = q.Encode()

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	c := &City{}
	err = json.NewDecoder(resp.Body).Decode(&c)
	return c, err
}
