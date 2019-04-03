package cache

import (
	"reflect"
	"testing"
	"time"
)

func TestSimpleCache(t *testing.T) {
	for _, c := range []struct {
		key   string
		value string
		ttl   time.Duration
		want  interface{}
	}{
		{
			"expired", "some string", 1 * time.Minute, nil,
		},
		{
			"cached", "some other string", 10 * time.Minute, []byte(`"some other string"`),
		},
	} {
		t.Run(c.key, func(t *testing.T) {
			r := &Simple{
				M: make(map[string]Value),
			}

			now = func() time.Time {
				return time.Date(2018, 02, 06, 11, 0, 0, 0, time.UTC)
			}

			if err := r.Put(c.key, c.value, c.ttl); err != nil {
				t.Fatal(err)
			}

			now = func() time.Time {
				return time.Date(2018, 02, 06, 11, 9, 0, 0, time.UTC)
			}

			got, err := r.Get(c.key)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v; want: %v", got, c.want)
			}
		})
	}
}
