package frontend

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/jarcoal/httpmock"
	"github.com/jonesrussell/jivesearch/log"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
)

func TestProxyHeaderHandler(t *testing.T) {
	for _, c := range []struct {
		name string
		q    string
		want *response
	}{
		{
			"basic", "https://example.com",
			&response{
				status:   http.StatusOK,
				template: "proxy_header",
				data: proxyResponse{
					Brand: Brand{
						Name:    "Some Name",
						TagLine: "A great tagline",
					},
					URL: "https://example.com",
				},
				err: nil,
			},
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			f := &Frontend{
				Brand: Brand{
					Name:    "Some Name",
					TagLine: "A great tagline",
				},
			}

			req, err := http.NewRequest("GET", "/proxy_header", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("q", c.q)
			req.URL.RawQuery = q.Encode()

			got := f.proxyHeaderHandler(httptest.NewRecorder(), req)

			if got.err != c.want.err {
				t.Fatalf("got error %v; want error %v", got.err, c.want.err)
			}
		})
	}
}

func TestProxyHandler(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		css    string
		q      string
		key    string
		secret string
		resp   string
	}

	for _, c := range []struct {
		name string
		args
		want string
	}{
		{
			"basic",
			args{
				css:    "",
				q:      "https://example.com",
				key:    "jfsdijf89sd",
				secret: "my_secret",
				resp: `<html>
								<head>
									<link rel="stylesheet" type="text/css" href="mystyle.css">
									<style>
										.body {margin:0}
										#mydiv {background: lightblue url("img_tree.gif") no-repeat fixed center}
									</style>									
									<script>alert("this is dangerous!")</script>
								</head>
								<body>
									<form id="form"></form>
									<a href="https://www.example.com">A link</a>
									<a href="/relative/link">A relative link</a>
									<iframe src="https://example.com/iframe/stuff"></iframe>
									<img src="nice.jpg" alt="nice image">
									<img src="data:image/png;base64, iVBORw0KGgoAAAANSUhEUgAAAggg==" alt="Red dot" />
									<div style="background-image: url('paper.gif');">Cool div you got there. Would be a shame if we proxied the url.</div>
								</body>
							</html>`,
			},
			`<html>
				<head>
				<link rel=stylesheet href="/proxy?css=true&amp;key=3jUnkmdp2GQ0a9mmkFWYaTq6pg9rxGdVlic5t4fvfKc%3D&amp;q=https%3A%2F%2Fexample.com%2Fmystyle.css">
				<style>
					.body {margin:0}
					#mydiv {background: lightblue url("/image/,sUKGG_QlTynjPRkhces2ykv26GkZbya3NOhrjgMZCWXY=/https://example.com/img_tree.gif") no-repeat fixed center}
				</style>
				</head>
				<body>
					<form id=form disabled action=javascript:void(0);></form>
					<a href="/proxy?key=_Zbla8JTucVtfb7n-QIGsrKozkTGaGsuKlxppnXb6xM%3D&amp;q=https%3A%2F%2Fwww.example.com" target=_top>A link</a>
					<a href="/proxy?key=j_gIsLDElFG1Qnp3TAYn1KD5dwvJ0gB_KqvUjXvM64g%3D&amp;q=https%3A%2F%2Fexample.com%2Frelative%2Flink" target=_top>A relative link</a>
					<iframe src="/proxy?iframe=true&amp;key=QtzD41Rkf5VUsmVPv9kSn4VHfUqf2jMljGktkjYVOVc%3D&amp;q=https%3A%2F%2Fexample.com%2Fiframe%2Fstuff"></iframe>
					<img src="/image/,sypKZuwtHssDFg_bLaExLhx4rYNnbr0KkzPeekQYRlGA=/https://example.com/nice.jpg" alt="nice image">
					<img src=data:image/png;base64 alt="Red dot">
					<div style='background-image:url("/image/,s1aMOcTAkBGs07NYeV9NjCCrDMIAQ7vtELioY-qfeDpo=/https://example.com/paper.gif")'>Cool div you got there. Would be a shame if we proxied the url.</div>
				</body>
			</html>`,
		},
		/*
			{
				"css",
				args{
					css:    "https://example.com/my.css",
					u:      "",
					key:    "jfsdijf89sd",
					secret: "my_secret",
					resp: `.body {margin:0}
								#mydiv {background-image: url(paper.gif)}
					`,
				},
				`.body {margin:0}
					#mydiv {background-image:url(/image/,s1aMOcTAkBGs07NYeV9NjCCrDMIAQ7vtELioY-qfeDpo=/https://example.com/paper.gif)}
				`,
			},
		*/
	} {
		t.Run(c.name, func(t *testing.T) {
			log.Debug.SetOutput(os.Stdout)

			f := &Frontend{
				Brand:       Brand{},
				ProxyClient: &http.Client{},
			}

			hmacSecret = func() string { return c.args.secret }
			k := hmacKey(c.q)

			responder := httpmock.NewStringResponder(200, c.resp)

			if c.args.q != "" {
				httpmock.RegisterResponder("GET", c.args.q, responder)
			}

			if c.args.css != "" {
				k = hmacKey(c.css)
				httpmock.RegisterResponder("GET", c.args.css, responder)
			}

			req, err := http.NewRequest("GET", "/proxy", nil)
			if err != nil {
				t.Fatal(err)
			}

			q := req.URL.Query()
			q.Add("css", c.css)
			q.Add("q", c.q)
			q.Add("key", k)
			req.URL.RawQuery = q.Encode()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(c.want))
			if err != nil {
				t.Fatal(err)
			}

			s, err := doc.Html()
			if err != nil {
				t.Fatal(err)
			}

			s, err = htmlMinify(s)
			if err != nil {
				t.Fatal(err)
			}

			got := f.proxyHandler(httptest.NewRecorder(), req)

			want := &response{
				status:   http.StatusOK,
				template: "proxy",
				data: proxyResponse{
					Brand: Brand{},
					HTML:  s,
					URL:   c.args.q,
				},
			}

			if c.args.q != "" {
				g := got.data.(proxyResponse)
				g.HTML, err = htmlMinify(g.HTML)
				if err != nil {
					t.Fatal(err)
				}

				got.data = g
			}

			if c.args.css != "" {
				want.template = "proxy_css"
				want.data = s

				g := got.data.(string)
				g, err = htmlMinify(g)
				if err != nil {
					t.Fatal(err)
				}

				got.data = g
			}

			if got.err != want.err {
				t.Fatalf("got error %v; want error %v", got.err, want.err)
			}
		})
	}

	httpmock.Reset()
}

func htmlMinify(s string) (string, error) {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/css", css.Minify)
	return m.String("text/html", s)
}
