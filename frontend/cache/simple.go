package cache

import (
	"encoding/json"
	"time"
)

// Simple implements the Cacher interface
type Simple struct {
	M map[string]interface{}
}

// Get retrieves an item from redis
func (s *Simple) Get(key string) (interface{}, error) {
	return s.M[key], nil
}

// Put sets a redis key to value
// TODO: Make this an expiring key
func (s *Simple) Put(key string, value interface{}, ttl time.Duration) error {
	j, err := json.Marshal(value)
	if err != nil {
		return err
	}

	s.M[key] = j
	return nil
}
