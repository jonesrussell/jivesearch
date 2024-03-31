// Command frontend demonstrates how to run the web app
package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	foo "log"

	"github.com/jonesrussell/jivesearch/instant/breach"
	"github.com/jonesrussell/jivesearch/instant/congress"
	"github.com/jonesrussell/jivesearch/instant/nutrition"
	"github.com/jonesrussell/jivesearch/instant/status"
	"github.com/jonesrussell/jivesearch/instant/whois"

	"github.com/jonesrussell/jivesearch/instant/econ/gdp"

	"github.com/jonesrussell/jivesearch/instant/currency"
	"github.com/jonesrussell/jivesearch/instant/econ/population"
	"github.com/jonesrussell/jivesearch/instant/shortener"

	"time"

	tzz "github.com/evanoberholster/timezoneLookup"
	"github.com/jonesrussell/jivesearch/instant/location"
	"github.com/jonesrussell/jivesearch/instant/weather"

	"github.com/abursavich/nett"
	"github.com/garyburd/redigo/redis"
	"github.com/jonesrussell/jivesearch/bangs"
	"github.com/jonesrussell/jivesearch/config"
	"github.com/jonesrussell/jivesearch/frontend"
	"github.com/jonesrussell/jivesearch/frontend/cache"
	"github.com/jonesrussell/jivesearch/instant"
	"github.com/jonesrussell/jivesearch/instant/discography/musicbrainz"
	"github.com/jonesrussell/jivesearch/instant/parcel"
	"github.com/jonesrussell/jivesearch/instant/stackoverflow"
	"github.com/jonesrussell/jivesearch/instant/stock"
	"github.com/jonesrussell/jivesearch/instant/timezone"
	"github.com/jonesrussell/jivesearch/instant/wikipedia"
	"github.com/jonesrussell/jivesearch/log"
	"github.com/jonesrussell/jivesearch/search"
	"github.com/jonesrussell/jivesearch/search/document"
	img "github.com/jonesrussell/jivesearch/search/image"
	"github.com/jonesrussell/jivesearch/search/provider"
	"github.com/jonesrussell/jivesearch/suggest"
	"github.com/olivere/elastic/v7"
	"github.com/spf13/viper"
	"golang.org/x/text/language"
)

var (
	f *frontend.Frontend
)

func setup(v *viper.Viper) *http.Server {
	v.SetEnvPrefix("jivesearch")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetDefaults(v)

	f = &frontend.Frontend{
		Brand: frontend.Brand{
			Name:      v.GetString("brand.name"),
			Host:      v.GetString("server.host"),
			TagLine:   v.GetString("brand.tagline"),
			Logo:      v.GetString("brand.logo"),
			SmallLogo: v.GetString("brand.small_logo"),
		},
		Onion: v.GetString("onion"),
	}

	router := f.Router(v)

	return &http.Server{
		Addr:    ":" + strconv.Itoa(v.GetInt("frontend.port")),
		Handler: http.TimeoutHandler(router, 5*time.Second, "Sorry, we took too long to get back to you"),
	}
}

func main() {
	v := viper.New()
	s := setup(v)

	var client *elastic.Client
	var err error

	httpClient := &http.Client{
		Transport: &http.Transport{
			Dial: (&nett.Dialer{
				Resolver: &nett.CacheResolver{TTL: 10 * time.Minute},
				IPFilter: nett.DualStack,
			}).Dial,
			DisableKeepAlives: true,
		},
		Timeout: 3 * time.Second,
	}

	switch v.GetString("search.provider") {
	case "yandex":
		f.Search = &provider.Yandex{
			Client: httpClient,
			Key:    v.GetString("yandex.key"),
			User:   v.GetString("yandex.user"),
		}
	default:
		f.Search = &search.ElasticSearch{
			ElasticSearch: &document.ElasticSearch{
				Client: esClient(v, client),
				Index:  v.GetString("elasticsearch.search.index"),
				Type:   v.GetString("elasticsearch.search.type"),
			},
		}
	}

	switch v.GetString("images.provider") {
	case "pixabay":
		f.Images.Fetcher = &img.Pixabay{
			HTTPClient: httpClient,
			Key:        v.GetString("pixabay.key"),
		}
	default:
		f.Images.Fetcher = &img.ElasticSearch{
			Client:        esClient(v, client),
			Index:         v.GetString("elasticsearch.images.index"),
			Type:          v.GetString("elasticsearch.images.type"),
			NSFWThreshold: .80,
		}
	}

	f.Images.Client = httpClient
	f.MapBoxKey = v.GetString("mapbox.key")

	// load naughty list
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	//parent := filepath.Dir(cwd)
	if err := suggest.NewNaughty(path.Join(cwd, "./suggest/naughty.txt")); err != nil {
		panic(err)
	}

	// !bangs
	debug := v.GetBool("debug")
	vb := viper.New()
	vb.SetConfigType("toml")
	vb.AddConfigPath("bangs")
	vb.SetConfigName("bangs") // the default !bangs config file
	if debug {
		vb.SetConfigName("bangs.test") // a shorter file to load quicker when debugging
	}

	// Print the current working directory
	cwd, err = os.Getwd()
	if err != nil {
		foo.Fatal(err)
	}
	fmt.Println("Current Working Directory:", cwd)

	// Assuming you've set the config path as shown in your code
	fmt.Println("Config Path:", "bangs")

	f.Bangs, err = bangs.New(vb)
	if err != nil {
		panic(err)
	}

	if err := f.Bangs.CreateFunctions(); err != nil {
		panic(err)
	}

	f.Cache.Instant = v.GetDuration("cache.instant")
	f.Cache.Search = v.GetDuration("cache.search")

	// The database needs to be setup beforehand.
	db, err := sql.Open("postgres",
		fmt.Sprintf(
			"user=%s password=%s host=%s database=%s sslmode=disable",
			v.GetString("postgresql.user"),
			v.GetString("postgresql.password"),
			v.GetString("postgresql.host"),
			v.GetString("postgresql.database"),
		),
	)
	if err != nil {
		panic(err)
	}

	defer db.Close()
	db.SetMaxIdleConns(0)

	// Instant Answers
	f.GitHub = frontend.GitHub{
		HTTPClient: httpClient,
	}

	f.Instant = &instant.Instant{
		QueryVar: "q",
		BreachFetcher: &breach.Pwned{
			HTTPClient: httpClient,
			UserAgent:  v.GetString("useragent"),
		},
		CongressFetcher: &congress.ProPublica{
			Key:        v.GetString("propublica.key"),
			HTTPClient: httpClient,
		},
		FedExFetcher: &parcel.FedEx{
			HTTPClient: httpClient,
			Account:    v.GetString("fedex.account"),
			Password:   v.GetString("fedex.password"),
			Key:        v.GetString("fedex.key"),
			Meter:      v.GetString("fedex.meter"),
		},
		Currency: instant.Currency{
			CryptoFetcher: &currency.CryptoCompare{
				Client:    httpClient,
				UserAgent: v.GetString("useragent"),
			},
			FXFetcher: &currency.ECB{},
		},
		GDPFetcher: &gdp.WorldBank{
			HTTPClient: httpClient,
		},
		LinkShortener: &shortener.IsGd{
			HTTPClient: httpClient,
		},
		NutritionFetcher: &nutrition.USDA{
			HTTPClient: httpClient,
			Key:        v.GetString("usda.key"),
		},
		PopulationFetcher: &population.WorldBank{
			HTTPClient: httpClient,
		},
		StackOverflowFetcher: &stackoverflow.API{
			HTTPClient: httpClient,
			Key:        v.GetString("stackoverflow.key"),
		},
		StatusFetcher: &status.IsItUp{
			HTTPClient: httpClient,
		},
		StockQuoteFetcher: &stock.IEX{
			HTTPClient: httpClient,
		},
		UPSFetcher: &parcel.UPS{
			HTTPClient: httpClient,
			User:       v.GetString("ups.user"),
			Password:   v.GetString("ups.password"),
			Key:        v.GetString("ups.key"),
		},
		USPSFetcher: &parcel.USPS{
			HTTPClient: httpClient,
			User:       v.GetString("usps.user"),
			Password:   v.GetString("usps.password"),
		},
		WeatherFetcher: &weather.OpenWeatherMap{
			HTTPClient: httpClient,
			Key:        v.GetString("openweathermap.key"),
		},
		WHOISFetcher: &whois.JiveData{ // until there are multiple whois fetchers Jive Data will be the default
			HTTPClient: httpClient,
			Key:        v.GetString("jivedata.key"),
		},
	}

	f.ProxyClient = httpClient

	// use Jive Data when debuggin to make setup easier
	switch debug {
	case true:
		log.Debug.SetOutput(os.Stdout)
		f.Cache.Cacher = &cache.Simple{
			M: make(map[string]cache.Value),
		}

		f.Bangs.Suggester = &bangs.Simple{}

		f.Suggest = &suggest.Simple{}

		f.Instant.DiscographyFetcher = &musicbrainz.JiveData{
			HTTPClient: httpClient,
			Key:        v.GetString("jivedata.key"),
		}

		f.Instant.LocationFetcher = &location.JiveData{
			HTTPClient: httpClient,
			Key:        v.GetString("jivedata.key"),
		}

		f.Instant.TimeZoneFetcher = &timezone.JiveData{
			HTTPClient: httpClient,
			Key:        v.GetString("jivedata.key"),
		}

		f.Instant.WikipediaFetcher = &wikipedia.JiveData{
			HTTPClient: httpClient,
			Key:        v.GetString("jivedata.key"),
		}
	default:
		// cache
		rds := &cache.Redis{
			RedisPool: &redis.Pool{
				MaxIdle:     1,
				MaxActive:   1,
				IdleTimeout: 10 * time.Second,
				Wait:        true,
				Dial: func() (redis.Conn, error) {
					cl, err := redis.Dial("tcp", fmt.Sprintf("%v:%v", v.GetString("redis.host"), v.GetString("redis.port")))
					if err != nil {
						return nil, err
					}
					return cl, err
				},
			},
		}

		defer rds.RedisPool.Close()

		f.Cache.Cacher = rds

		f.Bangs.Suggester = &bangs.ElasticSearch{
			Client: esClient(v, client),
			Index:  v.GetString("elasticsearch.bangs.index"),
			Type:   v.GetString("elasticsearch.bangs.type"),
		}

		f.Suggest = &suggest.ElasticSearch{
			Client: esClient(v, client),
			Index:  v.GetString("elasticsearch.query.index"),
			Type:   v.GetString("elasticsearch.query.type"),
		}

		f.Instant.DiscographyFetcher = &musicbrainz.PostgreSQL{
			DB: db,
		}

		f.Instant.LocationFetcher = &location.MaxMind{
			DB: v.GetString("maxmind.database"),
		}

		// timezone
		tz, err := tzz.LoadTimezones(tzz.Config{
			DatabaseType: "memory",
			DatabaseName: v.GetString("timezone.database"),
			Snappy:       true,
			Encoding:     tzz.EncJSON,
		})
		if err != nil {
			panic(err)
		}

		defer tz.Close()

		f.Instant.TimeZoneFetcher = &timezone.TZLookup{
			TZ: tz,
		}

		f.Instant.WikipediaFetcher = &wikipedia.PostgreSQL{
			DB: db,
		}
	}

	// setup !bangs suggester
	exists, err := f.Bangs.Suggester.IndexExists()
	if err != nil {
		panic(err)
	}

	if exists { // always want to recreate to add any changes/new !bangs
		if err := f.Bangs.Suggester.DeleteIndex(); err != nil {
			panic(err)
		}
	}

	if err := f.Bangs.Suggester.Setup(f.Bangs.Bangs); err != nil {
		panic(err)
	}

	// autocomplete & phrase suggestor
	exists, err = f.Suggest.IndexExists()
	if err != nil {
		panic(err)
	}

	if !exists {
		if err := f.Suggest.Setup(); err != nil {
			panic(err)
		}
	}

	// wikipedia setup
	if err := f.Instant.WikipediaFetcher.Setup(); err != nil {
		log.Info.Println(err)
	}

	// supported languages
	supported, unsupported := languages(v)
	for _, lang := range unsupported {
		log.Info.Printf("wikipedia does not support langugage %q\n", lang)
	}

	f.Wikipedia.Matcher = language.NewMatcher(supported)

	// see notes on customizing languages in search/document/document.go
	f.Document.Languages = document.Languages(supported)
	f.Document.Matcher = language.NewMatcher(f.Document.Languages)

	log.Info.Printf("Listening at http://127.0.0.1%v", s.Addr)
	log.Info.Fatal(s.ListenAndServe())
}

func esClient(v *viper.Viper, client *elastic.Client) *elastic.Client {
	if client == nil {
		var err error
		client, err = elastic.NewClient(
			elastic.SetURL(v.GetString("elasticsearch.url")),
			elastic.SetSniff(false),
		)

		if err != nil {
			panic(err)
		}
	}

	return client
}

func languages(cfg config.Provider) ([]language.Tag, []language.Tag) {
	supported := []language.Tag{}

	for _, l := range cfg.GetStringSlice("languages") {
		supported = append(supported, language.MustParse(l))
	}

	return wikipedia.Languages(supported)
}
