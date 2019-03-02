// Package config handles configuration settings
package config

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Provider outlines the configuration methods.
type Provider interface {
	SetDefault(key string, value interface{})
	SetTypeByDefaultValue(bool)
	BindPFlag(key string, flg *pflag.Flag) error
	Get(key string) interface{}
	GetString(key string) string
	GetInt(key string) int
	GetStringSlice(key string) []string
}

var now = func() time.Time { return time.Now().UTC() }

// SetDefaults configures some default values
func SetDefaults(cfg Provider) {
	cfg.SetTypeByDefaultValue(true)

	cfg.SetDefault("hmac.secret", "")

	// Brand
	cfg.SetDefault("brand.name", "Jive Search")
	cfg.SetDefault("brand.tagline", "A search engine that doesn't track you.")
	cfg.SetDefault("brand.logo",
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
		</svg>`)
	cfg.SetDefault("brand.small_logo",
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
		</svg>`)

	// Server
	port := 8000
	cfg.SetDefault("server.host", fmt.Sprintf("http://127.0.0.1:%d", port))

	// Frontend Cache
	cfg.SetDefault("cache.instant", 1*time.Second)
	cfg.SetDefault("cache.search", 1*time.Second)

	// languages are in the order of preference
	// empty slice = all languages
	// Note: the crawler and frontend packages (for now) don't support language config yet.
	// See note in search/document/document.go
	cfg.SetDefault("languages", []string{}) // e.g. JIVESEARCH_LANGUAGES="en fr de"

	// Elasticsearch
	cfg.SetDefault("elasticsearch.url", "http://127.0.0.1:9200")
	cfg.SetDefault("elasticsearch.search.index", "test-search")
	cfg.SetDefault("elasticsearch.search.type", "document")

	cfg.SetDefault("elasticsearch.bangs.index", "test-bangs")
	cfg.SetDefault("elasticsearch.bangs.type", "bang")

	cfg.SetDefault("elasticsearch.image.index", "test-images")
	cfg.SetDefault("elasticsearch.image.type", "image")

	cfg.SetDefault("elasticsearch.query.index", "test-queries")
	cfg.SetDefault("elasticsearch.query.type", "query")

	cfg.SetDefault("elasticsearch.robots.index", "test-robots")
	cfg.SetDefault("elasticsearch.robots.type", "robots")

	// PostgreSQL
	// Note: there is a security concern if postgres password is stored in env variable
	// but setting it as an env var w/in systemd nullifies this.
	cfg.SetDefault("postgresql.host", "localhost")
	cfg.SetDefault("postgresql.user", "jivesearch")
	cfg.SetDefault("postgresql.password", "mypassword")
	cfg.SetDefault("postgresql.database", "jivesearch")

	// Redis
	cfg.SetDefault("redis.host", "")
	cfg.SetDefault("redis.port", 6379)

	// crawler defaults
	tme := 5 * time.Minute
	cfg.SetDefault("crawler.useragent.full", "https://github.com/jivesearch/jivesearch")
	cfg.SetDefault("crawler.useragent.short", "jivesearchbot")
	cfg.SetDefault("crawler.time", tme.String())
	cfg.SetDefault("crawler.since", 30*24*time.Hour)
	cfg.SetDefault("crawler.seeds", []string{
		"https://moz.com/top500/domains",
		"https://domainpunch.com/tlds/topm.php",
		"https://www.wikipedia.org/",
	},
	)

	workers := 100
	cfg.SetDefault("crawler.workers", workers)
	cfg.SetDefault("crawler.max.bytes", 1024000) // 1MB...too little? too much??? Rememer <script> tags can take up a lot of bytes.
	cfg.SetDefault("crawler.timeout", 25*time.Second)
	cfg.SetDefault("crawler.max.queue.links", 100000)
	cfg.SetDefault("crawler.max.links", 100)
	cfg.SetDefault("crawler.max.domain.links", 10000)
	cfg.SetDefault("crawler.truncate.title", 100)
	cfg.SetDefault("crawler.truncate.keywords", 25)
	cfg.SetDefault("crawler.truncate.description", 250)

	// image nsfw scoring and metadata
	cfg.SetDefault("nsfw.host", "http://127.0.0.1:8080")
	cfg.SetDefault("nsfw.workers", 10)
	cfg.SetDefault("nsfw.since", now().AddDate(0, -1, 0))

	// Tor
	cfg.SetDefault("onion", "jivexx2rbi6llz37jq37n4uqff4kdipqbqd24c437c56om6uxbzhtdid.onion")

	// ProPublica API
	cfg.SetDefault("propublica.key", "my_key")

	// useragent for fetching api's, images, etc.
	cfg.SetDefault("useragent", "https://github.com/jivesearch/jivesearch")

	// stackoverflow API settings
	cfg.SetDefault("stackoverflow.key", "app key") // https://stackapps.com/apps/oauth/

	// FedEx package tracking API settings
	cfg.SetDefault("fedex.account", "account")
	cfg.SetDefault("fedex.password", "password")
	cfg.SetDefault("fedex.key", "key")
	cfg.SetDefault("fedex.meter", "meter")

	// Maps
	cfg.SetDefault("mapbox.key", "key")

	// MaxMind geolocation DB
	cfg.SetDefault("maxmind.database", "/usr/share/GeoIP/GeoLite2-City.mmdb")

	// Search Providers
	cfg.SetDefault("yandex.key", "key")
	cfg.SetDefault("yandex.user", "user")

	// UPS package tracking API settings
	cfg.SetDefault("ups.user", "user")
	cfg.SetDefault("ups.password", "password")
	cfg.SetDefault("ups.key", "key")

	// USPS package tracking API settings
	cfg.SetDefault("usps.user", "user")
	cfg.SetDefault("usps.password", "password")

	// OpenWeatherMap API settings
	cfg.SetDefault("openweathermap.key", "key")

	// Timezone database location
	cfg.SetDefault("timezone.database", "/usr/share/timezone/timezone") // suffix is automatically added

	// wikipedia settings
	truncate := 250
	cfg.SetDefault("wikipedia.truncate", truncate) // chars

	// command flags
	cmd := cobra.Command{}
	cmd.Flags().Int("workers", workers, "number of workers")
	cfg.BindPFlag("crawler.workers", cmd.Flags().Lookup("workers"))
	cmd.Flags().Duration("time", tme, "duration the crawler should run")
	cfg.BindPFlag("crawler.time", cmd.Flags().Lookup("time"))

	cmd.Flags().Int("port", port, "server port")
	cfg.BindPFlag("frontend.port", cmd.Flags().Lookup("port"))

	// control debug output
	cmd.Flags().Bool("debug", false, "turn on debug output")
	cfg.BindPFlag("debug", cmd.Flags().Lookup("debug"))

	// change search provider
	cmd.Flags().String("provider", "", "choose search provider")
	cfg.BindPFlag("search.provider", cmd.Flags().Lookup("provider"))

	// Jive Data
	cfg.SetDefault("jivedata.key", "TRIAL")
	cmd.Flags().Bool("jivedata", false, "use jivedata")
	cfg.BindPFlag("jivedata", cmd.Flags().Lookup("jivedata"))

	// wikipedia dump file settings
	cmd.Flags().String("dir", "", "path to save wiki dump files")
	cfg.BindPFlag("wikipedia.dir", cmd.Flags().Lookup("dir"))

	cmd.Flags().Bool("wikidata", true, "include wikidata")
	cfg.BindPFlag("wikipedia.wikidata", cmd.Flags().Lookup("wikidata"))

	cmd.Flags().Bool("wikipedia", true, "include wikipedia")
	cfg.BindPFlag("wikipedia.wikipedia", cmd.Flags().Lookup("wikipedia"))

	cmd.Flags().Bool("wikiquote", true, "include wikiquote")
	cfg.BindPFlag("wikipedia.wikiquote", cmd.Flags().Lookup("wikiquote"))

	cmd.Flags().Bool("wiktionary", true, "include wiktionary")
	cfg.BindPFlag("wikipedia.wiktionary", cmd.Flags().Lookup("wiktionary"))

	cmd.Flags().Int("truncate", truncate, "number of characters to extract from text")
	cfg.BindPFlag("wikipedia.truncate", cmd.Flags().Lookup("truncate"))

	cmd.Flags().Bool("delete", true, "delete file after parsed")
	cfg.BindPFlag("wikipedia.delete", cmd.Flags().Lookup("delete"))

	cfg.BindPFlag("wikipedia.workers", cmd.Flags().Lookup("workers"))

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
