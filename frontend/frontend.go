// Package frontend provides the routing and middleware for the web app
package frontend

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/jivesearch/jivesearch/bangs"
	"github.com/jivesearch/jivesearch/frontend/cache"
	"github.com/jivesearch/jivesearch/instant"
	"github.com/jivesearch/jivesearch/log"
	"github.com/jivesearch/jivesearch/search"
	img "github.com/jivesearch/jivesearch/search/image"
	"github.com/jivesearch/jivesearch/suggest"
	"github.com/oxtoacart/bpool"
	"golang.org/x/text/language"
)

// Frontend holds settings for branding, cache, search backend, etc.
type Frontend struct {
	Brand
	Document
	*bangs.Bangs
	Cache struct {
		cache.Cacher
		Instant time.Duration
		Search  time.Duration
	}
	Images struct {
		img.Fetcher
		*http.Client
	}
	*instant.Instant
	MapBoxKey   string
	Onion       string
	ProxyClient *http.Client
	Suggest     suggest.Suggester
	Search      search.Fetcher
	Wikipedia
	GitHub
}

// Brand allows for customization of the name and tagline
type Brand struct {
	Name      string
	Host      string
	TagLine   string
	Logo      string
	SmallLogo string
}

// Document has the languages we support
type Document struct {
	Languages []language.Tag
	language.Matcher
}

// Wikipedia holds our settings for wikipedia/wikidata
// Note: language matcher here may be different than that for
// document due to available languages Wikipedia supports
type Wikipedia struct {
	language.Matcher
}

var (
	bufpool   *bpool.BufferPool // makes sure no errors when writing to our templates
	templates map[string]*template.Template
)

func init() {
	bufpool = bpool.NewBufferPool(48) // what is the appropriate size??? 48??? 64???
}

type response struct {
	status   int
	redirect string
	template string
	data     interface{}
	err      error
}

type appHandler func(http.ResponseWriter, *http.Request) *response

// middleware sets a timeout and then serves.
func (f *Frontend) middleware(next appHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (fn appHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if rsp := fn(w, r); rsp != nil {
		switch rsp.status {
		case http.StatusOK:
			buf := bufpool.Get()
			defer bufpool.Put(buf)

			switch rsp.template {
			case "json":
				w.Header().Set("Content-Type", "application/json") // the default for json is utf-8
				err := json.NewEncoder(buf).Encode(rsp.data)
				if err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			case "jsonp":
				w.Header().Set("Content-Type", "application/javascript")
				err := json.NewEncoder(buf).Encode(rsp.data)
				if err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}

				fmt.Fprintf(w, "jivesearchcallback(%s)", buf)
				return // return here as we're done!
			case "proxy_css":
				w.Header().Set("Content-Type", "text/css; charset=utf-8")

				b := []byte(rsp.data.(string))

				if _, err := buf.Write(b); err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			case "proxy_iframe":
				b := []byte(rsp.data.(string))

				if _, err := buf.Write(b); err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			default: // parse the template
				switch rsp.template {
				case "opensearch":
					w.Header().Set("Content-Type", "application/opensearchdescription+xml")
				default:
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
				}

				tmpl, ok := templates[rsp.template]
				if !ok {
					rsp.status = http.StatusInternalServerError
					rsp.err = fmt.Errorf("template doesn't exist: %q", rsp.template)
					errHandler(w, rsp)
					return
				}

				if err := tmpl.Execute(buf, rsp.data); err != nil {
					rsp.status, rsp.err = http.StatusInternalServerError, err
					errHandler(w, rsp)
					return
				}
			}

			if _, err := buf.WriteTo(w); err != nil {
				rsp.status, rsp.err = http.StatusInternalServerError, err
				errHandler(w, rsp)
			}
		case http.StatusFound:
			switch rsp.data.(type) {
			case map[string][]string: // POST request
				// The http spec indicates 3xx redirects cannot change the
				// method (e.g. GET to POST). Thus, the code below with
				// http.Redirect will not work.
				/*
					m := rsp.data.(map[string][]string)
					j, err := json.Marshal(m)
					if err != nil {
						log.Info.Println(err)
					}

					r.Method = "POST"
					r.URL, _ = url.Parse(rsp.redirect)
					r.RequestURI = rsp.redirect
					r.Body = ioutil.NopCloser(bytes.NewReader(j))
					r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
					http.Redirect(w, r, rsp.redirect, http.StatusFound)
				*/
			default: // !bang
				http.Redirect(w, r, rsp.redirect, http.StatusFound)
			}
		case http.StatusBadRequest, http.StatusInternalServerError:
			errHandler(w, rsp)
		default:
			log.Info.Printf("Unknown status %d\n", rsp.status)
		}
	}
}

func errHandler(w http.ResponseWriter, rsp *response) {
	switch rsp.status {
	case http.StatusBadRequest:
		log.Debug.Println(rsp.err)
	case http.StatusInternalServerError:
		log.Info.Println(rsp.err)
	}

	http.Error(w, http.StatusText(rsp.status), rsp.status)
}

func (f *Frontend) autocompleteHandler(w http.ResponseWriter, r *http.Request) *response {
	var proxyFavIcon = func(u string) string {
		return fmt.Sprintf("/image/32x,s%v/%v", hmacKey(u), u)
	}

	q := strings.TrimSpace(r.FormValue("q"))

	if q == "!" {
		bngs := []bangs.Suggestion{}
		triggers := []string{"g", "a", "b", "reddit", "w"}
		for _, trigger := range triggers {
			for _, bng := range f.Bangs.Bangs {
				for _, tr := range bng.Triggers {
					if tr == trigger {
						sug := bangs.Suggestion{
							Trigger: trigger,
							Name:    bng.Name,
							FavIcon: proxyFavIcon(bng.FavIcon),
						}
						bngs = append(bngs, sug)
					}
				}
			}
		}

		// give a default set of !bang suggestions
		return &response{
			status:   http.StatusOK,
			template: "json",
			data: bangs.Results{
				Suggestions: bngs,
			},
		}

	} else if len(q) > 1 && !strings.HasPrefix(q, " ") && strings.HasPrefix(q, "!") {
		res, err := f.Bangs.Suggest(q, 10)
		if err != nil {
			return &response{
				status: http.StatusInternalServerError,
				err:    err,
			}
		}

		if len(res.Suggestions) > 0 {
			// fetch the favicons through our image proxy
			for i, b := range res.Suggestions {
				b.FavIcon = proxyFavIcon(b.FavIcon)
				res.Suggestions[i] = b
			}

			return &response{
				status:   http.StatusOK,
				template: "json",
				data:     res,
			}
		}
	}

	res, err := f.Suggest.Completion(q, 10)
	if err != nil {
		return &response{
			status: http.StatusInternalServerError,
			err:    err,
		}
	}

	return &response{
		status:   http.StatusOK,
		template: "json",
		data:     res,
	}
}

// ParseTemplates parses our html templates.
var ParseTemplates = func() {
	templates = make(map[string]*template.Template)
	templates["about"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"frontend/templates/base.html",
				"frontend/templates/search_form.html",
				"frontend/templates/about.html",
			),
	)

	var err error
	t := template.New("tmp")
	t, err = t.Parse(`{{template "answer" .}}`)
	if err != nil {
		panic(err)
	}

	templates["answer"] = template.Must(
		t.Funcs(funcMap).
			ParseFiles(
				"frontend/templates/wikipedia.html",
				"frontend/templates/answer.html",
			),
	)
	templates["maps"] = template.Must(
		template.New("maps.html").
			Funcs(funcMap).
			ParseFiles(
				"frontend/templates/maps.html",
			),
	)
	templates["opensearch"] = template.Must(
		template.New("opensearch.xml").
			ParseFiles(
				"frontend/templates/opensearch.xml",
			),
	)
	templates["proxy_header"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"frontend/templates/base.html",
				"frontend/templates/search_form.html",
				"frontend/templates/proxy_header.html",
			),
	)
	templates["proxy"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"frontend/templates/base.html",
				"frontend/templates/search_form.html",
				"frontend/templates/proxy.html",
			),
	)
	templates["search"] = template.Must(
		template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"frontend/templates/base.html",
				"frontend/templates/answer.html",
				"frontend/templates/search_form.html",
				"frontend/templates/search.html",
				"frontend/templates/wikipedia.html",
			),
	)
}
