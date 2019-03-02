package instant

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/instant/stock"
	"golang.org/x/text/language"
)

// StockQuoteType is an answer Type
const StockQuoteType Type = "stock quote"

// StockQuote is an instant answer
type StockQuote struct {
	Fetcher stock.Fetcher
	Answer
}

func (s *StockQuote) setQuery(r *http.Request, qv string) Answerer {
	s.Answer.setQuery(r, qv)
	return s
}

func (s *StockQuote) setUserAgent(r *http.Request) Answerer {
	return s
}

func (s *StockQuote) setLanguage(lang language.Tag) Answerer {
	s.language = lang
	return s
}

func (s *StockQuote) setType() Answerer {
	s.Type = StockQuoteType
	return s
}

func (s *StockQuote) setRegex() Answerer {
	triggers := []string{
		"stock quote", "stock price", "stock chart", "chart", "quote", "price", "stock",
	}

	for i, tr := range triggers {
		triggers[i] = tr + "[s]?"
	}

	t := strings.Join(triggers, "|")
	ticker := `^[\$]?[a-zA-Z]{1,5}[\.]?[a-zA-Z]?` // e.g. BRK.A

	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s)?\s?(?P<remainder>%s)$`, t, ticker)))
	s.regex = append(s.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>%s)\s(?P<trigger>%s)?$`, ticker, t)))

	return s
}

func (s *StockQuote) solve(r *http.Request) Answerer {
	ticker := strings.ToUpper(strings.Replace(s.remainder, "$", "", -1))

	// this shouldn't be handled in the package but by the caller...will fix later...
	for _, skipped := range []string{clock, currentTime, timeIn, wTime} {
		if strings.ToLower(ticker) == skipped {
			s.Err = fmt.Errorf("skipping stock. prefer another answer")
			fmt.Println("returning...")
			return s
		}
	}

	resp, err := s.Fetcher.Fetch(ticker)
	if err != nil {
		s.Err = err
		return s
	}

	resp = resp.SortHistorical()

	s.Data.Solution = resp
	return s
}

func (s *StockQuote) tests() []test {
	location, _ := time.LoadLocation("America/New_York")

	tests := []test{
		{
			query: "AAPL quote",
			expected: []Data{
				{
					Type:      StockQuoteType,
					Triggered: true,
					Solution: &stock.Quote{
						Ticker:   "AAPL",
						Name:     "Apple Inc.",
						Exchange: stock.NASDAQ,
						Last: stock.Last{
							Price:         171.42,
							Time:          time.Unix(1522090355062/1000, 0).In(location),
							Change:        6.48,
							ChangePercent: 0.03929,
						},
						History: []stock.EOD{
							{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
							{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
						},
						Provider: stock.IEXProvider,
					},
				},
			},
		},
		{
			query: "brk.a", // test for lowercase and has "."
			expected: []Data{
				{
					Type:      StockQuoteType,
					Triggered: true,
					Solution: &stock.Quote{
						Ticker:   "BRK.A",
						Name:     "Berkshire Hathaway",
						Exchange: stock.NYSE,
						Last: stock.Last{
							Price:         171.42,
							Time:          time.Unix(1522090355062/1000, 0).In(location),
							Change:        6.48,
							ChangePercent: 0.03929,
						},
						History: []stock.EOD{
							{Date: time.Date(2013, 3, 26, 0, 0, 0, 0, time.UTC), Open: 60.5276, Close: 59.9679, High: 60.5797, Low: 59.8891, Volume: 73428208},
							{Date: time.Date(2013, 3, 27, 0, 0, 0, 0, time.UTC), Open: 59.3599, Close: 58.7903, High: 59.4041, Low: 58.6147, Volume: 81854409},
						},
						Provider: stock.IEXProvider,
					},
				},
			},
		},
	}

	return tests
}
