package whois

import (
	"net/http"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
)

func TestJiveDataFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	for _, tt := range []struct {
		name string
		raw  string
	}{
		{
			name: "google.com",
			raw: `{
				"domain": "google.com",
				"domain_id": "2138514_DOMAIN_COM-VRSN",
				"status": "registered",
				"registrar": {
					"id": "292",
					"name": "MarkMonitor, Inc.",
					"organization": "MarkMonitor, Inc.",
					"url": "http://www.markmonitor.com"
				},
				"nameservers": [
					{
						"name": "ns3.google.com",
						"ipv4": null,
						"ipv6": null
					},
					{
						"name": "ns4.google.com",
						"ipv4": null,
						"ipv6": null
					},
					{
						"name": "ns1.google.com",
						"ipv4": null,
						"ipv6": null
					},
					{
						"name": "ns2.google.com",
						"ipv4": null,
						"ipv6": null
					}
				],
				"available": false,
				"registered": true,
				"created": "1997-09-15T00:00:00.000-07:00",
				"updated": "2018-02-21T10:45:07.000-08:00",
				"expires": "2020-09-13T21:00:00.000-07:00"
			}`,
		},
		{
			name: "somethingunregistered.com",
			raw: `{
				"domain": "somethingunregistered.com",
				"domain_id": "",
				"status": "available",
				"registrar": {
					"id": "",
					"name": "",
					"organization": "",
					"url": ""
				},
				"nameservers": [],
				"available": true,
				"registered": false,
				"created": "",
				"updated": "",
				"expires": ""
			}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			j := &JiveData{
				HTTPClient: &http.Client{},
				Key:        "somefakekey",
			}

			u, err := url.Parse("https://jivedata.com/whois")
			if err != nil {
				t.Fatal(err)
			}

			q := u.Query()
			q.Set("key", j.Key)
			q.Set("domain", tt.name)
			u.RawQuery = q.Encode()

			responder := httpmock.NewStringResponder(200, tt.raw)
			httpmock.RegisterResponder("GET", u.String(), responder)

			got, err := j.Fetch(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			want, err := wantResponse(tt.name)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}

	httpmock.Reset()
}

func wantResponse(domain string) (*Response, error) {
	var err error
	resp := &Response{}

	switch domain {
	case "google.com":
		resp = &Response{
			Domain:   domain,
			DomainID: "2138514_DOMAIN_COM-VRSN",
			Status:   "registered",
			Nameservers: []NameServer{
				{Name: "ns3.google.com"}, {Name: "ns4.google.com"}, {Name: "ns1.google.com"}, {Name: "ns2.google.com"},
			},
			Available:  false,
			Registered: true,
		}

		resp.Registrar.ID = "292"
		resp.Registrar.Name = "MarkMonitor, Inc."
		resp.Registrar.Organization = "MarkMonitor, Inc."
		resp.Registrar.URL = "http://www.markmonitor.com"

		created, err := time.Parse(time.RFC3339, "1997-09-15T00:00:00.000-07:00")
		if err != nil {
			return nil, err
		}
		resp.Created = MyTime{created}

		updated, err := time.Parse(time.RFC3339, "2018-02-21T10:45:07.000-08:00")
		if err != nil {
			return nil, err
		}
		resp.Updated = MyTime{updated}

		expires, err := time.Parse(time.RFC3339, "2020-09-13T21:00:00.000-07:00")
		if err != nil {
			return nil, err
		}
		resp.Expires = MyTime{expires}
	case "somethingunregistered.com":
		resp = &Response{
			Domain:      domain,
			Status:      "available",
			Nameservers: []NameServer{},
			Available:   true,
			Registered:  false,
		}
	}

	return resp, err
}
