package bangs

import (
	"strings"

	ferret "github.com/argusdusty/Ferret"
)

// Simple is a simple autocomplete suggester
type Simple struct {
	db *ferret.InvertedSuffix
}

// SuggestResults handles autocomplete queries
func (s *Simple) SuggestResults(term string, size int) (Results, error) {
	term = strings.TrimPrefix(term, "!")

	res := Results{}
	suggestions, _ := s.db.Query(term, size)

	for _, t := range suggestions {
		sug := Suggestion{
			Trigger: t,
		}

		res.Suggestions = append(res.Suggestions, sug)
	}

	return res, nil
}

// IndexExists returns true if the index exists
func (s *Simple) IndexExists() (bool, error) {
	return false, nil
}

// DeleteIndex will delete the existing index
func (s *Simple) DeleteIndex() error {
	return nil
}

// Setup recreates the completion index
func (s *Simple) Setup(bangs []Bang) error {
	s.db = ferret.New([]string{}, []string{}, []interface{}{}, func(s string) []byte { return []byte(s) })

	for _, b := range bangs {
		for _, t := range b.Triggers {
			s.db.Insert(t, t, []uint64{uint64(len(t))})
		}
	}

	return nil
}
