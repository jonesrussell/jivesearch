package provider

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jonesrussell/jivesearch/log"
	"github.com/jonesrussell/jivesearch/search"
	"github.com/jonesrussell/jivesearch/search/document"
	"golang.org/x/text/language"
)

// YandexProvider indicates the search results came from the Yandex API
var YandexProvider search.Provider = "Yandex"

// Yandex holds settings for the Yandex API.
type Yandex struct {
	*http.Client
	User string
	Key  string
}

// Fetch retrieves search results from the Yandex API.
// https://tech.yandex.com/xml/doc/dg/concepts/get-request-docpage/
// https://xml.yandex.com/test/
func (y *Yandex) Fetch(q string, filter search.Filter, lang language.Tag, region language.Region, number int, offset int) (*search.Results, error) {
	page := (offset / number) + 1

	u, err := y.buildYandexURL(q, filter, region, number, page)
	if err != nil {
		return nil, err
	}

	resp, err := y.Client.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	yr := &YandexResponse{}

	if err = xml.Unmarshal(bdy, &yr); err != nil {
		return nil, err
	}

	var ye error
	if yr.Response.Error != nil {
		ye = fmt.Errorf("yandex API error (%v): %q", yr.Response.Error.Code, yr.Response.Error.Text)
	}

	res := &search.Results{
		Provider: YandexProvider,
		Err:      ye,
	}

	for _, r := range yr.Response.Results.Grouping.Group {
		d, err := document.New(r.Doc.URL)
		if err != nil {
			log.Debug.Println(err)
			continue
		}

		d.Title = r.Doc.Title.Title
		d.Description = r.Doc.Headline.Headline
		if d.Description == "" {
			for _, p := range r.Doc.Passages.Passage {
				d.Description = p.Passage
				break
			}
		}

		d.Title = strings.Replace(d.Title, `<hlword>`, "", -1)
		d.Title = strings.Replace(d.Title, `</hlword>`, "", -1)
		d.Description = strings.Replace(d.Description, `<hlword>`, "", -1)
		d.Description = strings.Replace(d.Description, `</hlword>`, "", -1)

		res.Documents = append(res.Documents, d)

		for _, f := range yr.Response.Found {
			if f.Attrpriority == "all" {
				res.Count = f.Found
			}
		}
	}

	return res, err
}

// https://tech.yandex.com/xml/doc/dg/concepts/get-request-docpage/
func (y *Yandex) buildYandexURL(query string, filter search.Filter, region language.Region, number int, page int) (*url.URL, error) {
	u, err := url.Parse("https://yandex.com/search/xml")
	if err != nil {
		return nil, err
	}

	q := u.Query()
	q.Add("user", y.User)
	q.Add("key", y.Key)
	q.Add("query", query)
	//q.Add("lr", region.String()) // ID of the search country/region...only applies to Russian and Turkey search types
	q.Add("l10n", "en") // notification language
	//q.Add("sortby", "rlv") // relevancy by default
	q.Add("filter", string(filter))
	//q.Add("maxpassages", "")
	q.Add("groupby", fmt.Sprintf("attr=d.mode=deep.groups-on-page=%v.docs-in-group=1", number))
	q.Add("page", strconv.Itoa(page-1))
	q.Add("showmecaptcha", "no")

	u.RawQuery = q.Encode()
	log.Debug.Println(u.String())
	return u, err
}

// YandexResponse is the request and XML response from the Yandex API
type YandexResponse struct {
	Attrversion string `xml:"version,attr"  json:",omitempty"`
	Request     struct {
		Groupings struct {
			GroupBy struct {
				Attrattr                 string `xml:"attr,attr"  json:",omitempty"`
				Attrcurcateg             string `xml:"curcateg,attr"  json:",omitempty"`
				AttrdocsDashInDashGroup  string `xml:"docs-in-group,attr"  json:",omitempty"`
				AttrgroupsDashOnDashPage string `xml:"groups-on-page,attr"  json:",omitempty"`
				Attrmode                 string `xml:"mode,attr"  json:",omitempty"`
			} `xml:"groupby,omitempty" json:"groupby,omitempty"`
		} `xml:"groupings,omitempty" json:"groupings,omitempty"`
		MaxPassages string `xml:"maxpassages,omitempty" json:"maxpassages,omitempty"`
		Page        string `xml:"page,omitempty" json:"page,omitempty"`
		Query       string `xml:"query,omitempty" json:"query,omitempty"`
		SortBy      struct {
			Order    string `xml:"order,attr"  json:",omitempty"`
			Priority string `xml:"priority,attr"  json:",omitempty"`
			SortBy   string `xml:",chardata" json:"sortby,omitempty"`
		} `xml:"sortby,omitempty" json:"sortby,omitempty"`
	} `xml:"request,omitempty" json:"request,omitempty"`
	Response *Response `xml:"response,omitempty" json:"response,omitempty"`
}

// Response is the XML response from the Yandex API
type Response struct {
	Attrdate string `xml:"date,attr"  json:",omitempty"`
	Found    []struct {
		Attrpriority string `xml:"priority,attr"  json:",omitempty"`
		Found        int64  `xml:",chardata" json:",omitempty"`
	} `xml:"found,omitempty" json:"found,omitempty"`
	FoundHuman string `xml:"found-human,omitempty" json:"found-human,omitempty"`
	Reqid      string `xml:"reqid,omitempty" json:"reqid,omitempty"`
	Error      *struct {
		Text string `xml:",chardata"`
		Code string `xml:"code,attr"`
	} `xml:"error"`
	Results struct {
		Grouping struct {
			Attrattr                 string `xml:"attr,attr"  json:",omitempty"`
			Attrcurcateg             string `xml:"curcateg,attr"  json:",omitempty"`
			AttrdocsDashInDashGroup  string `xml:"docs-in-group,attr"  json:",omitempty"`
			AttrgroupsDashOnDashPage string `xml:"groups-on-page,attr"  json:",omitempty"`
			Attrmode                 string `xml:"mode,attr"  json:",omitempty"`
			Found                    []struct {
				Attrpriority string `xml:"priority,attr"  json:",omitempty"`
				Found        string `xml:",chardata" json:",omitempty"`
			} `xml:"found,omitempty" json:"found,omitempty"`
			FoundDocs []struct {
				Attrpriority string `xml:"priority,attr"  json:",omitempty"`
				FoundDocs    string `xml:",chardata" json:",omitempty"`
			} `xml:"found-docs,omitempty" json:"found-docs,omitempty"`
			FoundDocsHuman struct {
				FoundDocsHuman string `xml:",chardata" json:",omitempty"`
			} `xml:"found-docs-human,omitempty" json:"found-docs-human,omitempty"`
			Group []struct {
				Categ struct {
					Attrattr string `xml:"attr,attr"  json:",omitempty"`
					Attrname string `xml:"name,attr"  json:",omitempty"`
				} `xml:"categ,omitempty" json:"categ,omitempty"`
				Doc struct {
					Attrid   string `xml:"id,attr"  json:",omitempty"`
					Charset  string `xml:"charset,omitempty" json:"charset,omitempty"`
					Domain   string `xml:"domain,omitempty" json:"domain,omitempty"`
					Headline struct {
						Headline string `xml:",innerxml" json:",omitempty"`
					} `xml:"headline,omitempty" json:"headline,omitempty"`
					MimeType string `xml:"mime-type,omitempty" json:"mime-type,omitempty"`
					Modtime  string `xml:"modtime,omitempty" json:"modtime,omitempty"`
					Passages struct {
						Passage []struct {
							Passage string `xml:",innerxml" json:",omitempty"`
						} `xml:"passage,omitempty" json:"passage,omitempty"`
					} `xml:"passages,omitempty" json:"passages,omitempty"`
					Properties struct {
						TurboCgiURL   string `xml:"TurboCgiUrl,omitempty" json:"TurboCgiUrl,omitempty"`
						TurboFallback string `xml:"TurboFallback,omitempty" json:"TurboFallback,omitempty"`
						TurboLink     string `xml:"TurboLink,omitempty" json:"TurboLink,omitempty"`
						PassagesType  string `xml:"_PassagesType,omitempty" json:"_PassagesType,omitempty"`
						Lang          string `xml:"lang,omitempty" json:"lang,omitempty"`
					} `xml:"properties,omitempty" json:"properties,omitempty"`
					Relevance            string `xml:"relevance,omitempty" json:"relevance,omitempty"`
					SavedDashCopyDashURL string `xml:"saved-copy-url,omitempty" json:"saved-copy-url,omitempty"`
					Size                 string `xml:"size,omitempty" json:"size,omitempty"`
					Title                struct {
						Title string `xml:",innerxml" json:",omitempty"`
					} `xml:"title,omitempty" json:"title,omitempty"`
					URL string `xml:"url,omitempty" json:"url,omitempty"`
				} `xml:"doc,omitempty" json:"doc,omitempty"`
				DocCount  string `xml:"doccount,omitempty" json:"doccount,omitempty"`
				Relevance string `xml:"relevance,omitempty" json:"relevance,omitempty"`
			} `xml:"group,omitempty" json:"group,omitempty"`
			Page struct {
				Attrfirst string `xml:"first,attr"  json:",omitempty"`
				Attrlast  string `xml:"last,attr"  json:",omitempty"`
				Page      string `xml:",chardata" json:",omitempty"`
			} `xml:"page,omitempty" json:"page,omitempty"`
		} `xml:"grouping,omitempty" json:"grouping,omitempty"`
	} `xml:"results,omitempty" json:"results,omitempty"`
}
