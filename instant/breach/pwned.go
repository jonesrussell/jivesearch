package breach

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

// Pwned holds settings for the haveibeenpwned.com API
type Pwned struct {
	HTTPClient *http.Client
	UserAgent  string
}

// HaveIBeenPwnedProvider indicates the source is haveibeenpwned.com
const HaveIBeenPwnedProvider provider = "Have I Been Pwned"

// Fetch retrieves security breaches from haveibeenpwned.com
func (p *Pwned) Fetch(account string) (*Response, error) {
	u, err := p.buildURL(account)
	if err != nil {
		return nil, err
	}

	req, _ := http.NewRequest("GET", u.String(), nil)
	req.Header.Set("User-Agent", p.UserAgent)
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	pr := PwnedResponse{}

	if err := json.Unmarshal(bdy, &pr); err != nil {
		return nil, err
	}

	if len(pr) == 0 {
		return nil, err
	}

	r := New(account, HaveIBeenPwnedProvider)

	for _, b := range pr {
		d, err := time.Parse("2006-01-02", b.BreachDate)
		if err != nil {
			return nil, err
		}

		br := Breach{
			Name:        b.Name,
			Domain:      b.Domain,
			Date:        d,
			Count:       b.PwnCount,
			Description: b.Description,
			Items:       b.DataClasses,
		}

		r.Breaches = append(r.Breaches, br)
	}

	return r, err
}

func (p *Pwned) buildURL(account string) (*url.URL, error) {
	// https://haveibeenpwned.com/api/v2/breachedaccount/someone@example.com
	base, err := url.Parse("https://haveibeenpwned.com/")
	if err != nil {
		return nil, err
	}

	path, err := url.Parse(fmt.Sprintf("api/v2/breachedaccount/%v", account))
	if err != nil {
		return nil, err
	}

	u := base.ResolveReference(path)
	return u, err
}

// PwnedResponse is the raw response from haveibeenpwned.com
type PwnedResponse []struct {
	Title        string    `json:"Title"`
	Name         string    `json:"Name"`
	Domain       string    `json:"Domain"`
	BreachDate   string    `json:"BreachDate"`
	AddedDate    time.Time `json:"AddedDate"`
	ModifiedDate time.Time `json:"ModifiedDate"`
	PwnCount     int       `json:"PwnCount"`
	Description  string    `json:"Description"`
	DataClasses  []string  `json:"DataClasses"`
	IsVerified   bool      `json:"IsVerified"`
	IsFabricated bool      `json:"IsFabricated"`
	IsSensitive  bool      `json:"IsSensitive"`
	IsActive     bool      `json:"IsActive"`
	IsRetired    bool      `json:"IsRetired"`
	IsSpamList   bool      `json:"IsSpamList"`
	LogoType     string    `json:"LogoType"`
}
