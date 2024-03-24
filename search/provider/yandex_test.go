package provider

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/jivesearch/jivesearch/search"
	"github.com/jivesearch/jivesearch/search/document"
	"golang.org/x/text/language"
)

var YandexHendrixResponse = `<?xml version="1.0" encoding="utf-8"?>
<yandexsearch version="1.0">
    <request>
        <query>jimi hendrix</query>
        <page>0</page>
        <sortby order="descending" priority="no">rlv</sortby>
        <maxpassages></maxpassages>
        <groupings><groupby attr="d" mode="deep" groups-on-page="10" docs-in-group="1" curcateg="-1"/></groupings>
    </request>
    <response date="20180719T203659">
        <reqid>1532032619332668-1290890890466477902158225-man1-3616-XML</reqid>
        <found priority="phrase">6367388</found>
        <found priority="strict">6367388</found>
        <found priority="all">6367388</found>
        <found-human>6 mln. answers found</found-human>
        <results>
            <grouping attr="d" mode="deep" groups-on-page="10" docs-in-group="1" curcateg="-1">
                <found priority="phrase">9967</found>
                <found priority="strict">9967</found>
                <found priority="all">9967</found>
                <found-docs priority="phrase">5908619</found-docs>
                <found-docs priority="strict">5908619</found-docs>
                <found-docs priority="all">5908619</found-docs>
                <found-docs-human>found 6 mln. answers</found-docs-human>
                <page first="1" last="10">0</page>
                <group><categ attr="d" name="en.wikipedia.org"/>
                    <doccount>14993</doccount><relevance/>
                    <doc id="Z76637E6EFB3E0BD5"><relevance/>
                        <url>https://en.wikipedia.org/wiki/Jimi_Hendrix</url>
                        <domain>en.wikipedia.org</domain>
                        <title><hlword>Jimi Hendrix</hlword> - Wikipedia</title>
                        <headline>James Marshall (born Johnny Allen Hendrix)</headline>
                        <modtime>20070519T040000</modtime>
                        <size>27507</size>
                        <charset>utf-8</charset>
                        <properties>
                            <TurboCgiUrl>https://en.wikipedia.org/wiki/Jimi_Hendrix?utm_source=turbo</TurboCgiUrl>
                            <TurboFallback>https://en.wikipedia.org/wiki/Jimi_Hendrix</TurboFallback>
                            <TurboLink>https://yandex.ru/turbo?text=https%3A//en.wikipedia.org/wiki/Jimi_Hendrix%3Futm_source%3Dturbo</TurboLink>
                            <_PassagesType>0</_PassagesType>
                            <lang>en</lang>
                        </properties>
                        <mime-type>text/html</mime-type>
                        <saved-copy-url>https://hghltd.yandex.net/yandbtm?lang=en&amp;fmode=inject&amp;tm=1532032619&amp;tld=com&amp;la=1531516800&amp;text=jimi%20hendrix&amp;url=https%3A%2F%2Fen.wikipedia.org%2Fwiki%2FJimi_Hendrix&amp;l10n=en&amp;mime=html&amp;sign=3bc413a479bdae4d79e474debf1fc3cf&amp;keyno=0</saved-copy-url>
                    </doc>
                </group>
                <group><categ attr="d" name="jimihendrix.com"/>
                    <doccount>1845</doccount><relevance/>
                    <doc id="ZF2145B92CD33CFC5"><relevance/>
                        <url>http://www.jimihendrix.com/</url>
                        <domain>www.jimihendrix.com</domain>
                        <title><hlword>Jimi Hendrix</hlword> | The Official Site</title>
                        <modtime>20080113T030000</modtime>
                        <size>911</size>
                        <charset>utf-8</charset>
                        <passages>
                            <passage>Official Website of <hlword>Jimi Hendrix</hlword> with news, music, videos, album information and more!</passage>
                            <passage>SIGN UP FOR THE
                                <hlword>Jimi</hlword>
                                <hlword>Hendrix</hlword>
                                Newsletter.</passage>
                        </passages>
                        <properties>
                            <_PassagesType>0</_PassagesType>
                            <lang>en</lang>
                        </properties>
                        <mime-type>text/html</mime-type>
                        <saved-copy-url>http://hghltd.yandex.net/yandbtm?lang=en&amp;fmode=inject&amp;tm=1532032619&amp;tld=com&amp;la=1531621248&amp;text=jimi%20hendrix&amp;url=http%3A%2F%2Fwww.jimihendrix.com%2F&amp;l10n=en&amp;mime=html&amp;sign=b04931cf668bf6961c29bcb1374dee4a&amp;keyno=0</saved-copy-url>
                    </doc>
                </group>
            </grouping>
        </results>
    </response>
</yandexsearch>`

var YandexErrResponse = `<?xml version="1.0" encoding="utf-8"?><yandexsearch version="1.0"><response date="20190406T212117"><error code="33"> 90210.1.1.1 is not in this user&#x27;s list of permitted IP addresses</error><reqid>1</reqid></response></yandexsearch>`

func TestYandexFetch(t *testing.T) {
	type args struct {
		q      string
		filter search.Filter
		lang   language.Tag
		region language.Region
		number int
		page   int
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	doc1, err := document.New("https://en.wikipedia.org/wiki/Jimi_Hendrix")
	if err != nil {
		t.Fatal(err)
	}

	doc1.Title = "Jimi Hendrix - Wikipedia"
	doc1.Description = "James Marshall (born Johnny Allen Hendrix)"

	doc2, err := document.New("http://www.jimihendrix.com/")
	if err != nil {
		t.Fatal(err)
	}

	doc2.Title = "Jimi Hendrix | The Official Site"
	doc2.Description = "Official Website of Jimi Hendrix with news, music, videos, album information and more!"

	for _, tt := range []struct {
		name string
		args
		u     string
		yresp string
		want  *search.Results
	}{
		{
			name:  "moderate",
			args:  args{"jimi hendrix", search.Moderate, language.English, language.MustParseRegion("US"), 25, 1},
			u:     `https://yandex.com/search/xml?filter=moderate&groupby=attr%3Dd.mode%3Ddeep.groups-on-page%3D25.docs-in-group%3D1&key=key&l10n=en&page=0&query=jimi+hendrix&showmecaptcha=no&user=user`,
			yresp: YandexHendrixResponse,
			want: &search.Results{
				Provider: YandexProvider,
				Count:    6367388,
				Documents: []*document.Document{
					doc1, doc2,
				},
			},
		},
		{
			name:  "strict",
			args:  args{"jimi hendrix", search.Strict, language.English, language.MustParseRegion("US"), 25, 1},
			u:     `https://yandex.com/search/xml?filter=strict&groupby=attr%3Dd.mode%3Ddeep.groups-on-page%3D25.docs-in-group%3D1&key=key&l10n=en&page=0&query=jimi+hendrix&showmecaptcha=no&user=user`,
			yresp: YandexHendrixResponse,
			want: &search.Results{
				Provider: YandexProvider,
				Count:    6367388,
				Documents: []*document.Document{
					doc1, doc2,
				},
			},
		},
		{
			name:  "off",
			args:  args{"jimi hendrix", search.Off, language.English, language.MustParseRegion("US"), 25, 1},
			u:     `https://yandex.com/search/xml?filter=off&groupby=attr%3Dd.mode%3Ddeep.groups-on-page%3D25.docs-in-group%3D1&key=key&l10n=en&page=0&query=jimi+hendrix&showmecaptcha=no&user=user`,
			yresp: YandexHendrixResponse,
			want: &search.Results{
				Provider: YandexProvider,
				Count:    6367388,
				Documents: []*document.Document{
					doc1, doc2,
				},
			},
		},
		{
			name:  "error",
			args:  args{"jimi hendrix", search.Off, language.English, language.MustParseRegion("US"), 25, 1},
			u:     `https://yandex.com/search/xml?filter=off&groupby=attr%3Dd.mode%3Ddeep.groups-on-page%3D25.docs-in-group%3D1&key=key&l10n=en&page=0&query=jimi+hendrix&showmecaptcha=no&user=user`,
			yresp: YandexErrResponse,
			want: &search.Results{
				Provider: YandexProvider,
				Count:    0,
				Err:      fmt.Errorf(`yandex API error (33): " 90210.1.1.1 is not in this user's list of permitted IP addresses"`),
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(200, tt.yresp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			y := &Yandex{
				Client: &http.Client{},
				User:   "user",
				Key:    "key",
			}
			got, err := y.Fetch(tt.args.q, tt.args.filter, tt.args.lang, tt.args.region, tt.args.number, tt.args.page)
			if err != nil {
				t.Fatal(err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}

	httpmock.Reset()
}
