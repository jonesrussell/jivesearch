package cache

import (
	"encoding/json"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Redis implements the Cacher interface
type Redis struct {
	RedisPool *redis.Pool
	Prefix    string
}

func (r *Redis) prefixKey(key string) string {
	if r.Prefix == ""{
		r.Prefix = "jivesearch"	
	}
	return r.Prefix + "::" + key // jivesearch::key
}

// grab connection from pool and do the redis cmd
func (r *Redis) do(commandName string, args ...interface{}) (reply interface{}, err error) {
	c := r.RedisPool.Get()
	defer c.Close()

	return c.Do(commandName, args...)
}

// Get retrieves an item from redis
func (r *Redis) Get(key string) (interface{}, error) {
	key = r.prefixKey(key)
	return r.do("GET", key)
}

// Put sets a redis key to value
func (r *Redis) Put(key string, value interface{}, ttl time.Duration) error {
	s := seconds(ttl)

	j, err := json.Marshal(value)
	if err != nil {
		return err
	}

	key = r.prefixKey(key)
	ok, err := r.do("SET", key, j, "EX", s, "NX")
	if err != nil {
		return err
	}

	if ok != "OK" {
		err = ErrCannotSetKey
	}

	return err
}

func seconds(ttl time.Duration) int {
	return int(ttl / time.Second)
}
