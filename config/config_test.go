package config

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/spf13/pflag"
)

func TestSetDefaults(t *testing.T) {
	tme := 5 * time.Minute
	cfg := &provider{
		m: map[string]interface{}{},
	}

	now = func() time.Time {
		return time.Date(2018, 02, 06, 20, 34, 58, 651387237, time.UTC)
	}

	SetDefaults(cfg)

	port := 8000

	values := []struct {
		key   string
		value interface{}
	}{
		{"hmac.secret", ""},

		// Brand
		{"brand.name", "Jive Search"},
		{"brand.tagline", "A search engine that doesn't track you."},
		{"brand.logo",
			`<svg width="205" height="65" style="cursor:pointer;">
			<defs>
				<style>
					#logo {
						font-size: 36px;
						font-family: 'Open Sans',sans-serif;
						-webkit-touch-callout: none;
						-webkit-user-select: none;
						-khtml-user-select: none;
						-moz-user-select: none;
						-ms-user-select: none;
						user-select: none;
					}            
				</style>
			</defs>            
			<g><text id="logo" x="7" y="35" fill="#000">Jive Search</text></g>
		</svg>`},

		{"brand.small_logo",
			`<svg xmlns="http://www.w3.org/2000/svg" width="115px" height="48px">
			<defs>
				<style>
					#logo{
						font-size:20px;
					}            
				</style>
			</defs>
			<g>
				<text id="logo" x="0" y="37" fill="#000">Jive Search</text>
			</g>
		</svg>`},

		// Server
		{"server.host", fmt.Sprintf("http://127.0.0.1:%d", port)},

		// Elasticsearch
		{"elasticsearch.url", "http://127.0.0.1:9200"},
		{"elasticsearch.search.index", "test-search"},
		{"elasticsearch.search.type", "document"},
		{"elasticsearch.bangs.index", "test-bangs"},
		{"elasticsearch.bangs.type", "bang"},
		{"elasticsearch.image.index", "test-images"},
		{"elasticsearch.image.type", "image"},
		{"elasticsearch.query.index", "test-queries"},
		{"elasticsearch.query.type", "query"},
		{"elasticsearch.robots.index", "test-robots"},
		{"elasticsearch.robots.type", "robots"},

		// PostgreSQL
		{"postgresql.host", "localhost"},
		{"postgresql.user", "jivesearch"},
		{"postgresql.password", "mypassword"},
		{"postgresql.database", "jivesearch"},

		// Redis
		{"redis.host", ""},
		{"redis.port", 6379},

		// crawler defaults
		{"crawler.useragent.full", "https://github.com/jivesearch/jivesearch"},
		{"crawler.useragent.short", "jivesearchbot"},
		{"crawler.time", tme.String()},
		{"crawler.since", 30 * 24 * time.Hour},
		{"crawler.seeds", []string{
			"https://moz.com/top500/domains",
			"https://domainpunch.com/tlds/topm.php",
			"https://www.wikipedia.org/"},
		},
		{"crawler.workers", 100},
		{"crawler.max.bytes", 1024000},
		{"crawler.timeout", 25 * time.Second},
		{"crawler.max.queue.links", 100000},
		{"crawler.max.links", 100},
		{"crawler.max.domain.links", 10000},
		{"crawler.truncate.title", 100},
		{"crawler.truncate.keywords", 25},
		{"crawler.truncate.description", 250},

		// useragent for fetching api's, images, etc.
		{"useragent", "https://github.com/jivesearch/jivesearch"},

		// image nsfw scoring and metadata
		{"nsfw.workers", 10},
		{"nsfw.since", time.Date(2018, 01, 06, 20, 34, 58, 651387237, time.UTC)},

		// Tor
		{"onion", "jivexx2rbi6llz37jq37n4uqff4kdipqbqd24c437c56om6uxbzhtdid.onion"},

		// ProPublica API
		{"propublica.key", "my_key"},

		// stackoverflow API settings
		{"stackoverflow.key", "app key"},

		// FedEx package tracking API settings
		{"fedex.account", "account"},
		{"fedex.password", "password"},
		{"fedex.key", "key"},
		{"fedex.meter", "meter"},

		// Maps
		{"mapbox.key", "key"},

		// MaxMind geolocation DB
		{"maxmind.database", "/usr/share/GeoIP/GeoLite2-City.mmdb"},

		// Search Providers
		{"yandex.key", "key"},
		{"yandex.user", "user"},

		// UPS package tracking API settings
		{"ups.user", "user"},
		{"ups.password", "password"},
		{"ups.key", "key"},

		// USDA National Nutrient Database
		{"usda.key", "DEMO_KEY"},

		// USPS package tracking API settings
		{"usps.user", "user"},
		{"usps.password", "password"},

		// OpenWeatherMap API settings
		{"openweathermap.key", "key"},

		// Timezone database location
		{"timezone.database", "/usr/share/timezone/timezone"},

		// wikipedia settings
		{"wikipedia.truncate", 250},
	}

	for _, v := range values {
		got := cfg.Get(v.key)
		if !reflect.DeepEqual(got, v.value) {
			t.Fatalf("key %q; got %+v; want %+v", v.key, got, v.value)
		}
	}
}

type provider struct {
	m map[string]interface{}
}

func (p *provider) SetDefault(key string, value interface{}) {
	p.m[key] = value
}

func (p *provider) SetTypeByDefaultValue(bool) {}

func (p *provider) BindPFlag(key string, flg *pflag.Flag) error {
	return nil
}
func (p *provider) Get(key string) interface{} {
	return p.m[key]
}
func (p *provider) GetString(key string) string { return "" }
func (p *provider) GetInt(key string) int       { return 0 }
func (p *provider) GetStringSlice(key string) []string {
	return p.m[key].([]string)
}
