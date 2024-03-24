package suggest

import (
	ferret "github.com/argusdusty/Ferret"
)

// Simple is a simple autocomplete suggester
type Simple struct {
	db  *ferret.InvertedSuffix
	all []string
}

// Completion handles autocomplete queries
func (s *Simple) Completion(term string, size int) (Results, error) {
	res := Results{}
	res.Suggestions, _ = s.db.Query(term, size)
	return res, nil
}

// Exists checks if a term is already in our index
func (s *Simple) Exists(term string) (bool, error) {
	exists := false
	for _, w := range s.all {
		if w == term {
			exists = true
		}
	}

	return exists, nil
}

// Insert adds a new term to our index
func (s *Simple) Insert(term string) error {
	s.all = append(s.all, term)
	s.db.Insert(term, term, []uint64{uint64(len(term))})
	return nil
}

// Increment increments a term in our index
func (s *Simple) Increment(term string) error {
	return nil
}

// Setup creates a completion index
func (s *Simple) Setup() error {
	s.db = ferret.New([]string{}, []string{}, []interface{}{}, func(s string) []byte { return []byte(s) })
	return nil
}

// IndexExists returns true if the index exists
func (s *Simple) IndexExists() (bool, error) {
	return false, nil
}
