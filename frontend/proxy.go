package frontend

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/jivesearch/jivesearch/log"
)

type proxyResponse struct {
	Brand
	Context `json:"-"`
	HTML    string `json:"-"`
	URL     string `json:"-"`
}

func (f *Frontend) proxyHeaderHandler(w http.ResponseWriter, r *http.Request) *response {
	resp := &response{
		status:   http.StatusOK,
		template: "proxy_header",
		data: proxyResponse{
			Brand: f.Brand,
			URL:   r.FormValue("q"),
		},
		err: nil,
	}

	return resp
}

func (f *Frontend) proxyHandler(w http.ResponseWriter, r *http.Request) *response {
	u := r.FormValue("q")

	resp := &response{
		status:   http.StatusOK,
		template: "proxy",
		data: proxyResponse{
			Brand: f.Brand,
			URL:   u,
		},
		err: nil,
	}

	if u == "" {
		return resp
	}

	signature := r.FormValue("key")

	css := r.FormValue("css")
	if css == "true" {
		uu, err := url.Parse(u)
		if err != nil {
			log.Debug.Println(err)
			return resp
		}

		if !validSignature([]byte(hmacSecret()), uu, signature) {
			return resp
		}

		res, err := f.get(uu)
		if err != nil {
			log.Debug.Println(err)
			return resp
		}

		defer res.Body.Close()

		h, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Debug.Println(err)
			return resp
		}

		resp.data = replaceCSS(uu, string(h))

		resp.template = "proxy_css"
		return resp
	}

	base, err := url.Parse(u)
	if err != nil {
		log.Debug.Println(err)
		return resp
	}

	if !validSignature([]byte(hmacSecret()), base, signature) {
		return resp
	}

	res, err := f.get(base)
	if err != nil {
		log.Debug.Println(err)
		return resp
	}

	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Debug.Println(err)
		return resp
	}

	// TODO: remove all comments...no need for them

	// remove all javascript
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		s.Remove()
	})

	// disable all forms
	doc.Find("form").Each(func(i int, s *goquery.Selection) {
		s.SetAttr("disabled", "disabled") // doesn't seem to work....
		s.SetAttr("action", "javascript:void(0);")
	})

	// proxy links
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		for _, href := range []string{"href"} {
			if lnk, ok := s.Attr(href); ok {
				u, err := createProxyLink(base, lnk, false)
				if err != nil {
					log.Debug.Println(err)
					return
				}

				s.SetAttr(href, u.String())
				s.SetAttr("target", "_top") // make all links open in the main page (not w/in the iframe)
			}
		}
	})

	// proxy iframe src
	doc.Find("iframe").Each(func(i int, s *goquery.Selection) {
		if lnk, ok := s.Attr("src"); ok {
			u, err := createProxyLink(base, lnk, true)
			if err != nil {
				log.Debug.Println(err)
				return
			}

			s.SetAttr("src", u.String())
		}
	})

	// proxy images
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		for _, src := range []string{"src", "srcset"} {
			if lnk, ok := s.Attr(src); ok {
				if lnk == "" {
					continue
				}

				matches := reSrcSet.FindAllStringSubmatch(lnk, -1)
				if len(matches) == 0 {
					u := createProxyImage(base, lnk)
					s.SetAttr(src, u)
					continue
				}

				lnks := []string{}

				for _, m := range matches {
					if isBase64(m[1]) {
						lnks = append(lnks, m[1])
						continue
					}

					u := createProxyImage(base, m[1])

					lnks = append(lnks, fmt.Sprintf("%v %v", u, m[2]))
				}

				s.SetAttr(src, strings.Join(lnks, " "))
			}
		}
	})

	// proxy url() within <style></style> tags
	doc.Find("style").Each(func(i int, s *goquery.Selection) {
		h := replaceCSS(base, s.Text())
		s.ReplaceWithHtml(fmt.Sprintf(`<style>%v</style>`, h))
	})

	// proxy url() for any inline styles
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		attr := "style"
		if style, ok := s.Attr(attr); ok {
			s.SetAttr(attr, replaceCSS(base, style))
		}
	})

	// replace external stylesheets with a proxied version
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		if rel, ok := s.Attr("rel"); ok && strings.ToLower(rel) == "stylesheet" {
			if lnk, ok := s.Attr("href"); ok {
				u, err := createProxyCSSLink(base, lnk)
				if err != nil {
					log.Debug.Println(err)
					return
				}

				s.SetAttr("href", u.String())
			}
		}
	})

	h, err := doc.Html()
	if err != nil {
		log.Debug.Println(err)
		return resp
	}

	switch r.FormValue("iframe") {
	case "true":
		resp.template = "proxy_iframe"
		resp.data = string(h)
	default:
		resp.data = proxyResponse{
			Brand: f.Brand,
			HTML:  h,
			URL:   u,
		}
	}

	return resp
}

func isBase64(s string) bool {
	return strings.HasPrefix(strings.ToLower(s), "data:")
}

// can have ', ", or no quotes
var reCSSLinkReplacer = regexp.MustCompile(`(url\(['"]?)(?P<link>.*?)['"]?\)`)
var reSrcSet = regexp.MustCompile(`(?P<url>.*?),? (?P<size>[0-9]+(\.[0-9]+)?[wx],?)?\s?`)

// https://stackoverflow.com/a/28005189/522962
func replaceAllSubmatchFunc(re *regexp.Regexp, b []byte, f func(s []byte) []byte) []byte {
	idxs := re.FindAllSubmatchIndex(b, -1)
	if len(idxs) == 0 {
		return b
	}
	l := len(idxs)
	ret := append([]byte{}, b[:idxs[0][0]]...)
	for i, pair := range idxs {
		ret = append(ret, f(b[pair[4]:pair[5]])...) // 2 & 3 are <url>. 4 & 5 are the <link>
		if i+1 < l {
			ret = append(ret, b[pair[1]:idxs[i+1][0]]...)
		}
	}
	ret = append(ret, b[idxs[len(idxs)-1][1]:]...)
	return ret
}

func replaceCSS(base *url.URL, s string) string {
	// replace any urls with a proxied link
	ss := replaceAllSubmatchFunc(reCSSLinkReplacer, []byte(s), func(ss []byte) []byte {
		if isBase64(string(ss)) { // base64 image
			return []byte(fmt.Sprintf("url(%q)", ss))
		}

		m := string(ss)

		u, err := url.Parse(m)
		if err != nil {
			log.Info.Println(err)
		}

		u = base.ResolveReference(u)

		key := hmacKey(u.String())
		uu := fmt.Sprintf("/image/,s%v/%v", key, u.String())
		return []byte(fmt.Sprintf("url(%q)", uu))
	})

	return string(ss)
}

func createProxyImage(base *url.URL, lnk string) string {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)
	key := hmacKey(u.String())
	l := fmt.Sprintf("/image/,s%v/%v", key, u.String())
	return l
}

func createProxyCSSLink(base *url.URL, lnk string) (*url.URL, error) {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)

	uu, err := url.Parse("/proxy")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Add("key", hmacKey(u.String()))
	q.Add("q", u.String())
	q.Add("css", "true")
	uu.RawQuery = q.Encode()
	return uu, err
}

func createProxyLink(base *url.URL, lnk string, iframe bool) (*url.URL, error) {
	u, err := url.Parse(lnk)
	if err != nil {
		panic(err)
	}

	u = base.ResolveReference(u)

	uu, err := url.Parse("/proxy")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Add("key", hmacKey(u.String()))
	q.Add("q", u.String())
	if iframe {
		q.Add("iframe", "true")
	}
	uu.RawQuery = q.Encode()
	return uu, err
}

// validSignature returns whether the request signature is valid.
func validSignature(key []byte, u *url.URL, signature string) bool {
	if m := len(signature) % 4; m != 0 { // add padding if missing
		signature += strings.Repeat("=", 4-m)
	}

	got, err := base64.URLEncoding.DecodeString(signature)
	if err != nil {
		log.Debug.Printf("error base64 decoding signature %q\n", signature)
		return false
	}

	mac := hmac.New(sha256.New, key)
	if _, err := mac.Write([]byte(u.String())); err != nil {
		log.Debug.Println(err)
		return false
	}
	want := mac.Sum(nil)

	return hmac.Equal(got, want)
}

func (f *Frontend) get(u *url.URL) (*http.Response, error) {
	// we don't want &httputil.ReverseProxy as we don't want to pass the user's IP address & other info.
	request, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	return f.ProxyClient.Do(request)
}
