// Package cache provides simple caching mechanisms
package cache

import (
	"errors"
	"time"
)

// ErrCannotSetKey indicates an issue setting the key
var ErrCannotSetKey = errors.New("unable to set key")

// Cacher outlines the methods for our cache
type Cacher interface {
	Get(key string) (interface{}, error)
	Put(key string, value interface{}, ttl time.Duration) error
}
