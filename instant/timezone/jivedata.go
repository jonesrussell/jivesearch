package timezone

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// JiveData holds settings for the Jive Data API
type JiveData struct {
	HTTPClient *http.Client
	Key        string
}

// Fetch retrieves the timezone from Jive Data
func (j *JiveData) Fetch(lat, lon float64) (string, error) {
	u, err := url.Parse("https://jivedata.com/timezone")
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("key", j.Key)
	q.Set("lat", strconv.FormatFloat(lat, 'f', -1, 64))
	q.Set("lon", strconv.FormatFloat(lon, 'f', -1, 64))
	u.RawQuery = q.Encode()

	resp, err := j.HTTPClient.Get(u.String())
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	r := struct {
		TimeZone string `json:"timezone"`
	}{}

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r.TimeZone, err
}
