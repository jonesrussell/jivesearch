// Package stackoverflow fetches stackoverflow data
package stackoverflow

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

// Fetcher retrieves a StackOverflowResponse
type Fetcher interface {
	Fetch(query string, tags []string) (Response, error)
}

// API retrieves information from the Stack Overflow API
type API struct {
	Key        string
	HTTPClient *http.Client
}

// Response is the raw response from Stack Overflow
type Response struct {
	Items          []Item `json:"items"`
	QuotaMax       int    `json:"quota_max"`
	QuotaRemaining int    `json:"quota_remaining"`
}

// Item is a question with answers, a link, and a title
type Item struct {
	Answers []Answer `json:"answers"`
	Link    string   `json:"link"`
	Title   string   `json:"title"`
}

// Answer is a single answer
type Answer struct {
	Owner `json:"owner"`
	Score int    `json:"score"`
	Body  string `json:"body"`
}

// Owner is the person who answered the question
// TODO: I wasn't able to get both the User's display name and link to their profile or id.
// Can select one or the other but not both in their filter.
type Owner struct {
	DisplayName string `json:"display_name"`
}

func (a *API) buildURL(query string, tags []string) (string, error) {
	// find the question
	// To edit fields returned, etc...
	// https://api.stackexchange.com/docs/advanced-search#page=1&pagesize=1&order=desc&sort=relevance&q=sum%20variables&answers=1&tagged=php&filter=!OfZYd4zGqhN8IapZI6RQ6uaya_ZCewDWcGt5p6k_N2q&site=stackoverflow&run=true
	u, err := url.Parse("https://api.stackexchange.com/2.2/search/advanced")
	if err != nil {
		return "", err
	}

	// We search questions (ranked by relevancy) with at least 1 answer.
	// We then take the highest upvoted answer, whether or not it is the accepted answer.
	// e.g. a search for "php loop": https://stackoverflow.com/search?q=%5Bphp%5D+loop
	// TODO: make params more configurable
	q := u.Query()
	q.Set("key", a.Key)
	q.Set("q", query)
	q.Set("tagged", strings.Join(tags, ","))
	q.Set("page", "1")
	q.Set("pagesize", "1")
	q.Set("order", "desc")
	q.Set("sort", "relevance")
	q.Set("answers", "1") // has at least 1 answer
	q.Set("site", "stackoverflow")
	q.Set("filter", "!OfZYd4zGqhN8IapZI6RQ6uaya_ZBeR7bHr1c)NV5Cu9")
	u.RawQuery = q.Encode()

	return u.String(), nil
}

// Fetch retrieves a Stack Overflow response
func (a *API) Fetch(query string, tags []string) (Response, error) {
	r := Response{}

	u, err := a.buildURL(query, tags)
	if err != nil {
		return r, err
	}

	resp, err := a.HTTPClient.Get(u)
	if err != nil {
		return r, err
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&r)
	return r, err
}
