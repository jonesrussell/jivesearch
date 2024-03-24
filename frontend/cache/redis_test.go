package cache

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/rafaeljusto/redigomock"
)

func TestGet(t *testing.T) {
	for _, c := range []struct {
		key   string
		value string
	}{
		{
			"first", "some string",
		},
		{
			"second", "some other string",
		},
	} {
		t.Run(c.key, func(t *testing.T) {
			r := &Redis{}
			conn := redigomock.NewConn()
			conn.Command("GET", r.prefixKey(c.key)).Expect(c.value)

			r.RedisPool = &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return conn, nil
				},
			}
			defer r.RedisPool.Close()

			got, err := r.Get(c.key)
			if err != nil {
				t.Fatal(err)
			}

			if got != c.value {
				t.Fatalf("got %v; want: %v", got, c.value)
			}
		})
	}
}

func TestPut(t *testing.T) {
	type test struct {
		Name string
	}

	for _, c := range []struct {
		key   string
		value interface{}
		ttl   time.Duration
	}{
		{
			"first", "some string", 10 * time.Minute,
		},
		{
			"second", test{"bob"}, 1 * time.Minute,
		},
	} {
		t.Run(c.key, func(t *testing.T) {
			j, err := json.Marshal(c.value)
			if err != nil {
				t.Fatal(err)
			}

			r := &Redis{}
			conn := redigomock.NewConn()
			conn.Command("SET", r.prefixKey(c.key), j, "EX", int(c.ttl/time.Second), "NX").Expect("OK")

			r.RedisPool = &redis.Pool{
				Dial: func() (redis.Conn, error) {
					return conn, nil
				},
			}
			defer r.RedisPool.Close()

			if err := r.Put(c.key, c.value, c.ttl); err != nil {
				t.Fatal(err)
			}
		})
	}
}
