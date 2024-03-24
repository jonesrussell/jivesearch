package cache

import (
	"encoding/json"
	"time"
)

// Simple implements the Cacher interface
type Simple struct {
	M map[string]Value
}

// Value is a value with an expiration
type Value struct {
	value   interface{}
	expires time.Time
}

var now = func() time.Time { return time.Now().UTC() }

// Get retrieves an item from redis
func (s *Simple) Get(key string) (interface{}, error) {
	if val, ok := s.M[key]; ok {
		if now().Before(val.expires) {
			return val.value, nil
		}
	}

	return nil, nil
}

// Put sets a redis key to value
func (s *Simple) Put(key string, val interface{}, ttl time.Duration) error {
	j, err := json.Marshal(val)
	if err != nil {
		return err
	}

	v := Value{
		value:   j,
		expires: now().Add(ttl),
	}

	s.M[key] = v
	return nil
}
