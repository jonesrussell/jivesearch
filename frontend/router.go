package frontend

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/jivesearch/jivesearch/config"
	"willnorris.com/go/imageproxy"
)

// Router sets up the routes & handlers
func (f *Frontend) Router(cfg config.Provider) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	router.NewRoute().Name("search").Methods("GET").Path("/").Handler(
		f.middleware(appHandler(f.searchHandler)),
	)
	router.NewRoute().Name("answer").Methods("GET").Path("/answer").Handler(
		f.middleware(appHandler(f.answerHandler)),
	)
	router.NewRoute().Name("about").Methods("GET").Path("/about").Handler(
		f.middleware(appHandler(f.aboutHandler)),
	)
	router.NewRoute().Name("autocomplete").Methods("GET").Path("/autocomplete").Handler(
		f.middleware(appHandler(f.autocompleteHandler)),
	)
	router.NewRoute().Name("favicon").Methods("GET").Path("/favicon.ico").Handler(
		http.FileServer(http.Dir("static")),
	)
	router.NewRoute().Name("opensearch").Methods("GET").Path("/opensearch.xml").Handler(
		f.middleware(appHandler(f.openSearchHandler)),
	)
	router.NewRoute().Name("proxy").Methods("GET").Path("/proxy").Handler(
		f.middleware(appHandler(f.proxyHandler)),
	)
	router.NewRoute().Name("proxy_header").Methods("GET").Path("/proxy_header").Handler(
		f.middleware(appHandler(f.proxyHeaderHandler)),
	)

	// How do we exclude viewing the entire static directory of /static path?
	router.NewRoute().Name("static").Methods("GET").PathPrefix("/static/").Handler(
		corsHeaders(http.StripPrefix("/static/", http.FileServer(http.Dir("static")))),
	)

	// make hmac key available to our templates
	key := cfg.GetString("hmac.secret")
	os.Setenv("hmac_secret", key)

	p := imageproxy.NewProxy(nil, nil)
	p.Verbose = false // otherwise logs the image fetched
	//p.UserAgent = cfg.GetString("useragent") // not implemented yet: https://github.com/willnorris/imageproxy/pull/83
	p.SignatureKey = []byte(key)
	p.Timeout = 2 * time.Second
	router.NewRoute().Name("image").Methods("GET").PathPrefix("/image/").Handler(http.StripPrefix("/image", p))

	/* To generate new HMAC secret...
	// DON'T RUN IN PLAYGROUND! Will get same secret each time ;)
	b := make([]byte, 32) // s/b at least 32
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	fmt.Println(base64.URLEncoding.EncodeToString(b))
	*/

	return router
}

// for fonticons for Wikipedia widget
func corsHeaders(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
		h.ServeHTTP(w, r)
	})
}
