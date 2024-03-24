// Package whois fetchers WHOIS information for a domain
package whois

import (
	"time"
)

// Fetcher retrieves WHOIS information for a domain
type Fetcher interface {
	Fetch(domain string) (*Response, error)
}

type provider string

// Response is a WHOIS response
type Response struct {
	Domain    string `json:"domain" xml:"domain"`
	DomainID  string `json:"domain_id" xml:"domain_id"`
	Status    string `json:"status" xml:"status"`
	Registrar struct {
		ID           string `json:"id" xml:"id"`
		Name         string `json:"name" xml:"name"`
		Organization string `json:"organization" xml:"organization"`
		URL          string `json:"url" xml:"url"`
	} `json:"registrar" xml:"registrar"`
	RegistrantContacts []interface{} `json:"registrant_contacts,omitempty" xml:"registrant_contacts,omitempty"` // no longer available for .com's under GDPR?
	AdminContacts      []interface{} `json:"admin_contacts,omitempty" xml:"admin_contacts,omitempty"`           // no longer available for .com's under GDPR?
	TechnicalContacts  []interface{} `json:"technical_contacts,omitempty" xml:"technical_contacts,omitempty"`   // no longer available for .com's under GDPR?
	Nameservers        []NameServer  `json:"nameservers" xml:"nameservers"`
	Available          bool          `json:"available" xml:"available"`
	Registered         bool          `json:"registered" xml:"registered"`
	Created            MyTime        `json:"created" xml:"created"` // we need a custom unmarshaller for `""`
	Updated            MyTime        `json:"updated" xml:"updated"` // we need a custom unmarshaller for `""`
	Expires            MyTime        `json:"expires" xml:"expires"` // we need a custom unmarshaller for `""`
	Disclaimer         string        `json:"disclaimer" xml:"disclaimer"`
	Raw                string        `json:"raw" xml:"raw"`
	Error              string        `json:"error" xml:"error"`
}

// NameServer is a nameserver
type NameServer struct {
	Name string      `json:"name" xml:"name"`
	Ipv4 interface{} `json:"ipv4" xml:"ipv4"`
	Ipv6 interface{} `json:"ipv6" xml:"ipv6"`
}

// MyTime helps us unmarshal "" as time.Time
type MyTime struct {
	time.Time
}

// UnmarshalJSON is a patch for processing blank values
func (m *MyTime) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package. Add `""` for blank values
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	tt, err := time.Parse(`"`+time.RFC3339+`"`, string(data))
	*m = MyTime{tt}
	return err
}
