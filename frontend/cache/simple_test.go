package cache

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestSimpleCache(t *testing.T) {
	for _, c := range []struct {
		key   string
		value string
		ttl   time.Duration
	}{
		{
			"first", "some string", 1 * time.Minute,
		},
		{
			"second", "some other string", 10 * time.Minute,
		},
	} {
		t.Run(c.key, func(t *testing.T) {
			r := &Simple{
				M: make(map[string]interface{}),
			}

			if err := r.Put(c.key, c.value, c.ttl); err != nil {
				t.Fatal(err)
			}

			j, err := json.Marshal(c.value)
			if err != nil {
				t.Fatal(err)
			}

			got, err := r.Get(c.key)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, j) {
				t.Fatalf("got %v; want: %v", got, c.value)
			}
		})
	}
}
