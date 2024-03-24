// Package stock fetches stock quote data
package stock

import (
	"fmt"
	"sort"
	"time"
)

// Fetcher retrieves stock quotes
type Fetcher interface {
	Fetch(ticker string) (*Quote, error)
}

type provider string
type exchange string

// NYSE is the New York Stock Exchange
const NYSE exchange = "NYSE"

// NASDAQ is the NASDAQ Stock Exchange
const NASDAQ exchange = "NASDAQ"

// ErrInvalidTicker indicates the ticker was invalid
var ErrInvalidTicker = fmt.Errorf("invalid ticker")

// Quote includes the current and historical quotes
type Quote struct {
	Ticker   string
	Name     string
	Exchange exchange
	Last
	History  []EOD
	Provider provider
}

// Last is the latest quote
type Last struct {
	Price         float64
	Time          time.Time
	Change        float64
	ChangePercent float64
	// more to come
}

// EOD is an end of day stock quote
type EOD struct {
	Date   time.Time `json:"date"`
	Open   float64   `json:"open"`
	Close  float64   `json:"close"`
	High   float64   `json:"high"`
	Low    float64   `json:"low"`
	Volume int       `json:"volume"`
}

// get the Stock Exchange from a string
func (q *Quote) exchange(e string) (*Quote, error) {
	var err error

	switch e {
	case "Nasdaq Global Select":
		q.Exchange = NASDAQ
	case "New York Stock Exchange":
		q.Exchange = NYSE
	default:
		err = fmt.Errorf("unknown stock exchange %v", e)
	}

	return q, err
}

// SortHistorical sorts the historical quotes in ascending order
func (q *Quote) SortHistorical() *Quote {
	sort.Slice(q.History, func(i, j int) bool {
		return q.History[i].Date.Before(q.History[j].Date)
	})

	return q
}
