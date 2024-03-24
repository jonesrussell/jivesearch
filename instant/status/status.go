// Package status checks if a website is down
package status

import "golang.org/x/net/publicsuffix"

// Fetcher implements methods to check for checking a website status
type Fetcher interface {
	Fetch(domain string) (*Response, error)
}

type provider string

// Response is a status response
type Response struct {
	Domain   string  `json:"domain"`
	Port     int     `json:"port"`
	Status   int     `json:"status_code"`
	IP       string  `json:"response_ip"`
	Code     int     `json:"response_code"`
	Time     float64 `json:"response_time"`
	Provider provider
}

// FixDomain appends .com to a string if there is no tld
func FixDomain(domain string) string {
	tld, err := publicsuffix.EffectiveTLDPlusOne(domain)
	if err != nil { // we're just gonna guess if it needs a .com
		tld = domain + ".com"
	}

	return tld
}
