package frontend

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/spf13/pflag"
)

func TestRouter(t *testing.T) {
	type route struct {
		name   string
		method string
		url    string
	}

	for _, c := range []*route{
		{
			name:   "search",
			method: "GET",
			url:    "https://www.example.com/?q=search+term",
		},
		{
			name:   "answer",
			method: "GET",
			url:    "https://www.example.com/answer/?q=search+term",
		},
		{
			name:   "about",
			method: "GET",
			url:    "http://localhost/about",
		},
		{
			name:   "autocomplete",
			method: "GET",
			url:    "http://127.0.0.1/autocomplete",
		},
		{
			name:   "favicon",
			method: "GET",
			url:    "http://example.com/favicon.ico",
		},
		{
			name:   "static",
			method: "GET",
			url:    "https://example.com/static/main.js",
		},
		{
			name:   "opensearch",
			method: "GET",
			url:    "http://localhost/opensearch.xml",
		},
		{
			name:   "proxy",
			method: "GET",
			url:    "http://localhost/proxy/?u=https://example.com",
		},
		{
			name:   "proxy_header",
			method: "GET",
			url:    "http://localhost/proxy_header",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			cfg := &mockProvider{
				m: make(map[string]interface{}),
			}
			cfg.SetDefault("hmac.secret", "very secret")

			f := &Frontend{}
			router := f.Router(cfg)

			expected, err := http.NewRequest(
				c.method,
				c.url,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			route := router.Get(c.name)

			if !route.Match(expected, &mux.RouteMatch{}) {
				t.Fatalf("expected route for %q to exist. It doesn't", c.url)
			}
		})
	}
}

type mockProvider struct {
	m map[string]interface{}
}

func (p *mockProvider) SetDefault(key string, value interface{}) {
	p.m[key] = value
}
func (p *mockProvider) SetTypeByDefaultValue(bool) {}
func (p *mockProvider) BindPFlag(key string, flg *pflag.Flag) error {
	return nil
}
func (p *mockProvider) Get(key string) interface{} {
	return p.m[key]
}
func (p *mockProvider) GetString(key string) string {
	return p.m[key].(string)
}
func (p *mockProvider) GetInt(key string) int {
	return p.m[key].(int)
}
func (p *mockProvider) GetStringSlice(key string) []string {
	return p.m[key].([]string)
}
