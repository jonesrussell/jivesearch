// Package congress retrieves members of the United States Congress
package congress

import (
	"strings"
)

// ErrInvalidState indicates an invalid state was passed in
var ErrInvalidState error

// Location is the member's state
type Location struct {
	Short string
	State string
}

// Response is contains the members of Congress/Senate for a district/state
type Response struct {
	*Location
	Role
	Members []Member
	Provider
}

// Role is a congressional role
type Role string

const (
	// Senators are United States Senators
	Senators Role = "Senators"
	// House are United States House Members
	House Role = "House Members"
)

// Provider is a data source
type Provider string

// Member is a member of Congress
type Member struct {
	Name         string
	District     int
	Gender       string
	Party        string
	Twitter      string
	Facebook     string
	NextElection int
}

// Fetcher implements methods to retrieve members of Congress/Senate for a district/state
type Fetcher interface {
	FetchSenators(location *Location) (*Response, error)
	FetchMembers(location *Location) (*Response, error)
}

// ValidateState returns a valid code for a State
func ValidateState(state string) *Location {
	state = strings.ToLower(state)

	// did they pass in a code or the name of the state?
	for k, v := range usc {
		if k == state || v == state {
			l := &Location{
				Short: strings.ToUpper(v),
				State: strings.Title(k),
			}
			return l
		}
	}

	return nil
}

var usc = map[string]string{
	"alaska":         "ak",
	"alabama":        "al",
	"arkansas":       "ar",
	"arizona":        "az",
	"california":     "ca",
	"colorado":       "co",
	"connecticut":    "ct",
	"delaware":       "de",
	"florida":        "fl",
	"georgia":        "ga",
	"hawaii":         "hi",
	"iowa":           "ia",
	"idaho":          "id",
	"illinois":       "il",
	"indiana":        "in",
	"kansas":         "ks",
	"kentucky":       "ky",
	"louisiana":      "la",
	"massachusetts":  "ma",
	"maryland":       "md",
	"maine":          "me",
	"michigan":       "mi",
	"minnesota":      "mn",
	"missouri":       "mo",
	"mississippi":    "ms",
	"montana":        "mt",
	"north carolina": "nc",
	"north dakota":   "nd",
	"nebraska":       "ne",
	"new hampshire":  "nh",
	"new jersey":     "nj",
	"new mexico":     "nm",
	"nevada":         "nv",
	"new york":       "ny",
	"ohio":           "oh",
	"oklahoma":       "ok",
	"oregon":         "or",
	"pennsylvania":   "pa",
	"rhode island":   "ri",
	"south carolina": "sc",
	"south dakota":   "sd",
	"tennessee":      "tn",
	"texas":          "tx",
	"utah":           "ut",
	"virginia":       "va",
	"vermont":        "vt",
	"washington":     "wa",
	"wisconsin":      "wi",
	"west virginia":  "wv",
	"wyoming":        "wy",

	// Territories...e.g. non-voting members (https://en.wikipedia.org/wiki/Non-voting_members_of_the_United_States_House_of_Representatives)
	"american samoa":       "as",
	"district of dolumbia": "dc",
	//"federated states of micronesia": "fm",
	"guam": "gu",
	//"marshall islands":         "mh",
	"northern mariana islands": "mp",
	//"palau":                    "pw",
	"puerto rico":    "pr",
	"virgin islands": "vi",
}
