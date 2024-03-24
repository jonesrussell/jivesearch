// Package wikipedia fetches Wikipedia articles
package wikipedia

import (
	"encoding/json"
	"regexp"
	"strings"

	"golang.org/x/text/language"
)

// Fetcher outlines the methods used to retrieve Wikipedia snippets
type Fetcher interface {
	Setup() error
	Fetch(query string, lang language.Tag) ([]*Item, error)
}

// Item is the contains the complete wiki info for a person, thing or word.
type Item struct {
	Wikipedia
	*Wikidata
	Wikiquote  Wikiquote  `json:"wikiquote"`
	Wiktionary Wiktionary `json:"wiktionary"`
}

// Wikipedia holds the summary text of an article
type Wikipedia struct {
	ID           string   `json:"wikibase_item"`
	Language     string   `json:"language"`
	OutgoingLink []string `json:"outgoing_link,omitempty"`
	Popularity   float64  `json:"popularity_score,omitempty"`
	Title        string   `json:"title"`
	Text         string   `json:"text"`
	truncate     int
}

var reParen = regexp.MustCompile(`\s?\((.*?)\)`) // replace parenthesis

// UnmarshalJSON truncates the text
func (w *Wikipedia) UnmarshalJSON(data []byte) error {
	// copy the fields of Wikipedia but not the
	// methods so we don't recursively call UnmarshalJSON
	type Alias Wikipedia
	a := &struct {
		*Alias
	}{
		Alias: (*Alias)(w),
	}

	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}

	w.Text = reParen.ReplaceAllString(w.Text, "")
	w.Text = strings.Replace(w.Text, "\u00a0", "", -1) // otherwise causes a panic below

	if w.truncate != -1 && len(w.Text) > w.truncate { // truncates while preserving words.
		c := strings.Fields(w.Text[:w.truncate+1])
		w.Text = strings.Join(c[0:len(c)-1], " ")
		if !strings.HasSuffix(w.Text, ".") {
			w.Text = w.Text + " ..."
		}
	}

	return nil
}

// Languages verifies languages based on Wikipedia's supported languages.
// An empty slice of supported languages implies you support every language available.
func Languages(supported []language.Tag) ([]language.Tag, []language.Tag) {
	// make sure supported languages are supported by Wikipedia
	s := []language.Tag{}
	unsupported := []language.Tag{}

	switch len(supported) {
	case 0:
		for lang := range Available {
			s = append(s, lang)
		}
	default:
		for _, lang := range supported {
			if _, ok := Available[lang]; !ok {
				unsupported = append(unsupported, lang)
				continue
			}

			s = append(s, lang)
		}
	}

	return s, unsupported
}

// Available is a map of all languages that Wikipedia supports.
// https://en.wikipedia.org/wiki/List_of_Wikipedias
// There is also a separate entry for Wiktionary and Wikiquote:
//	 https://en.wiktionary.org/wiki/Wiktionary:List_of_languages
//	 https://en.wikiquote.org/wiki/Wikiquote:Other_language_Wikiquotes
// We sort their table by # of Articles descending.
var Available = map[language.Tag]struct{}{
	language.MustParse("en"):         {}, // english is fallback
	language.MustParse("ceb"):        {},
	language.MustParse("sv"):         {},
	language.MustParse("de"):         {},
	language.MustParse("nl"):         {},
	language.MustParse("fr"):         {},
	language.MustParse("ru"):         {},
	language.MustParse("it"):         {},
	language.MustParse("es"):         {},
	language.MustParse("war"):        {},
	language.MustParse("pl"):         {},
	language.MustParse("vi"):         {},
	language.MustParse("ja"):         {},
	language.MustParse("pt"):         {},
	language.MustParse("zh"):         {},
	language.MustParse("uk"):         {},
	language.MustParse("ca"):         {},
	language.MustParse("fa"):         {},
	language.MustParse("ar"):         {},
	language.MustParse("no"):         {},
	language.MustParse("sh"):         {},
	language.MustParse("fi"):         {},
	language.MustParse("hu"):         {},
	language.MustParse("id"):         {},
	language.MustParse("ro"):         {},
	language.MustParse("cs"):         {},
	language.MustParse("ko"):         {},
	language.MustParse("sr"):         {},
	language.MustParse("tr"):         {},
	language.MustParse("ms"):         {},
	language.MustParse("eu"):         {},
	language.MustParse("eo"):         {},
	language.MustParse("bg"):         {},
	language.MustParse("da"):         {},
	language.MustParse("min"):        {},
	language.MustParse("kk"):         {},
	language.MustParse("sk"):         {},
	language.MustParse("hy"):         {},
	language.MustParse("zh-min-nan"): {},
	language.MustParse("he"):         {},
	language.MustParse("lt"):         {},
	language.MustParse("hr"):         {},
	language.MustParse("ce"):         {},
	language.MustParse("sl"):         {},
	language.MustParse("et"):         {},
	language.MustParse("gl"):         {},
	language.MustParse("nn"):         {},
	language.MustParse("uz"):         {},
	language.MustParse("el"):         {},
	language.MustParse("be"):         {},
	language.MustParse("la"):         {},
	//language.MustParse("simple"):struct{}{}, // Simple English...does not parse
	language.MustParse("vo"):        {},
	language.MustParse("hi"):        {},
	language.MustParse("ur"):        {},
	language.MustParse("th"):        {},
	language.MustParse("az"):        {},
	language.MustParse("ka"):        {},
	language.MustParse("ta"):        {},
	language.MustParse("cy"):        {},
	language.MustParse("mk"):        {},
	language.MustParse("mg"):        {},
	language.MustParse("oc"):        {},
	language.MustParse("lv"):        {},
	language.MustParse("bs"):        {},
	language.MustParse("new"):       {},
	language.MustParse("tt"):        {},
	language.MustParse("tg"):        {},
	language.MustParse("te"):        {},
	language.MustParse("tl"):        {},
	language.MustParse("sq"):        {},
	language.MustParse("pms"):       {},
	language.MustParse("ky"):        {},
	language.MustParse("br"):        {},
	language.MustParse("be-tarask"): {},
	language.MustParse("zh-yue"):    {},
	language.MustParse("ht"):        {},
	language.MustParse("jv"):        {},
	language.MustParse("ast"):       {},
	language.MustParse("bn"):        {},
	language.MustParse("lb"):        {},
	language.MustParse("ml"):        {},
	language.MustParse("mr"):        {},
	language.MustParse("af"):        {},
	language.MustParse("pnb"):       {},
	language.MustParse("sco"):       {},
	language.MustParse("is"):        {},
	language.MustParse("ga"):        {},
	language.MustParse("cv"):        {},
	language.MustParse("ba"):        {},
	language.MustParse("fy"):        {},
	language.MustParse("sw"):        {},
	language.MustParse("my"):        {},
	language.MustParse("lmo"):       {},
	language.MustParse("an"):        {},
	language.MustParse("yo"):        {},
	language.MustParse("ne"):        {},
	language.MustParse("io"):        {},
	language.MustParse("gu"):        {},
	language.MustParse("nds"):       {},
	language.MustParse("scn"):       {},
	language.MustParse("bpy"):       {},
	language.MustParse("pa"):        {},
	language.MustParse("ku"):        {},
	language.MustParse("als"):       {},
	language.MustParse("bar"):       {},
	language.MustParse("kn"):        {},
	language.MustParse("qu"):        {},
	language.MustParse("ia"):        {},
	language.MustParse("su"):        {},
	language.MustParse("ckb"):       {},
	language.MustParse("mn"):        {},
	language.MustParse("arz"):       {},
	language.MustParse("bat-smg"):   {},
	language.MustParse("azb"):       {},
	language.MustParse("nap"):       {},
	language.MustParse("wa"):        {},
	language.MustParse("gd"):        {},
	language.MustParse("bug"):       {},
	language.MustParse("yi"):        {},
	language.MustParse("am"):        {},
	language.MustParse("map-bms"):   {},
	language.MustParse("si"):        {},
	language.MustParse("fo"):        {},
	language.MustParse("mzn"):       {},
	language.MustParse("or"):        {},
	language.MustParse("li"):        {},
	language.MustParse("sah"):       {},
	language.MustParse("hsb"):       {},
	language.MustParse("vec"):       {},
	language.MustParse("sa"):        {},
	language.MustParse("os"):        {},
	language.MustParse("mai"):       {},
	language.MustParse("ilo"):       {},
	language.MustParse("mrj"):       {},
	language.MustParse("hif"):       {},
	language.MustParse("mhr"):       {},
	language.MustParse("xmf"):       {},
	//language.MustParse("roa-tara"):struct{}{}, // Does not parse
	language.MustParse("nah"): {},
	//language.MustParse("eml"):struct{}{}, // Does not parse
	language.MustParse("bh"):      {},
	language.MustParse("pam"):     {},
	language.MustParse("ps"):      {},
	language.MustParse("nso"):     {},
	language.MustParse("diq"):     {},
	language.MustParse("hak"):     {},
	language.MustParse("sd"):      {},
	language.MustParse("se"):      {},
	language.MustParse("mi"):      {},
	language.MustParse("bcl"):     {},
	language.MustParse("nds-nl"):  {},
	language.MustParse("gan"):     {},
	language.MustParse("glk"):     {},
	language.MustParse("vls"):     {},
	language.MustParse("rue"):     {},
	language.MustParse("bo"):      {},
	language.MustParse("wuu"):     {},
	language.MustParse("szl"):     {},
	language.MustParse("fiu-vro"): {},
	language.MustParse("sc"):      {},
	language.MustParse("co"):      {},
	language.MustParse("vep"):     {},
	language.MustParse("lrc"):     {},
	language.MustParse("tk"):      {},
	language.MustParse("csb"):     {},
	//language.MustParse("zh-classical"):struct{}{}, // Does not parse
	language.MustParse("crh"):     {},
	language.MustParse("km"):      {},
	language.MustParse("gv"):      {},
	language.MustParse("kv"):      {},
	language.MustParse("frr"):     {},
	language.MustParse("as"):      {},
	language.MustParse("lad"):     {},
	language.MustParse("zea"):     {},
	language.MustParse("so"):      {},
	language.MustParse("cdo"):     {},
	language.MustParse("ace"):     {},
	language.MustParse("ay"):      {},
	language.MustParse("udm"):     {},
	language.MustParse("kw"):      {},
	language.MustParse("stq"):     {},
	language.MustParse("nrm"):     {},
	language.MustParse("ie"):      {},
	language.MustParse("lez"):     {},
	language.MustParse("myv"):     {},
	language.MustParse("koi"):     {},
	language.MustParse("rm"):      {},
	language.MustParse("pcd"):     {},
	language.MustParse("ug"):      {},
	language.MustParse("lij"):     {},
	language.MustParse("mt"):      {},
	language.MustParse("fur"):     {},
	language.MustParse("gn"):      {},
	language.MustParse("dsb"):     {},
	language.MustParse("gom"):     {},
	language.MustParse("dv"):      {},
	language.MustParse("cbk-zam"): {},
	language.MustParse("ext"):     {},
	language.MustParse("ang"):     {},
	language.MustParse("kab"):     {},
	language.MustParse("mwl"):     {},
	language.MustParse("ksh"):     {},
	language.MustParse("ln"):      {},
	language.MustParse("gag"):     {},
	language.MustParse("sn"):      {},
	language.MustParse("nv"):      {},
	language.MustParse("frp"):     {},
	language.MustParse("pag"):     {},
	language.MustParse("pi"):      {},
	language.MustParse("av"):      {},
	language.MustParse("lo"):      {},
	language.MustParse("dty"):     {},
	language.MustParse("xal"):     {},
	language.MustParse("pfl"):     {},
	language.MustParse("krc"):     {},
	language.MustParse("haw"):     {},
	language.MustParse("kaa"):     {},
	language.MustParse("olo"):     {},
	language.MustParse("bxr"):     {},
	language.MustParse("rw"):      {},
	language.MustParse("pdc"):     {},
	language.MustParse("pap"):     {},
	language.MustParse("bjn"):     {},
	language.MustParse("to"):      {},
	language.MustParse("nov"):     {},
	language.MustParse("kl"):      {},
	language.MustParse("arc"):     {},
	language.MustParse("jam"):     {},
	language.MustParse("kbd"):     {},
	language.MustParse("ha"):      {},
	language.MustParse("tet"):     {},
	language.MustParse("tyv"):     {},
	language.MustParse("tpi"):     {},
	language.MustParse("ki"):      {},
	language.MustParse("ig"):      {},
	language.MustParse("na"):      {},
	language.MustParse("ab"):      {},
	language.MustParse("lbe"):     {},
	language.MustParse("roa-rup"): {},
	language.MustParse("jbo"):     {},
	language.MustParse("ty"):      {},
	language.MustParse("kg"):      {},
	language.MustParse("za"):      {},
	language.MustParse("lg"):      {},
	language.MustParse("wo"):      {},
	language.MustParse("mdf"):     {},
	language.MustParse("srn"):     {},
	language.MustParse("zu"):      {},
	language.MustParse("bi"):      {},
	language.MustParse("ltg"):     {},
	language.MustParse("chr"):     {},
	language.MustParse("tcy"):     {},
	language.MustParse("sm"):      {},
	language.MustParse("om"):      {},
	language.MustParse("tn"):      {},
	language.MustParse("chy"):     {},
	language.MustParse("xh"):      {},
	language.MustParse("tw"):      {},
	language.MustParse("cu"):      {},
	language.MustParse("rmy"):     {},
	language.MustParse("tum"):     {},
	language.MustParse("pih"):     {},
	language.MustParse("rn"):      {},
	language.MustParse("got"):     {},
	language.MustParse("pnt"):     {},
	language.MustParse("ss"):      {},
	language.MustParse("ch"):      {},
	language.MustParse("bm"):      {},
	language.MustParse("ady"):     {},
	language.MustParse("mo"):      {},
	language.MustParse("ts"):      {},
	language.MustParse("iu"):      {},
	language.MustParse("st"):      {},
	language.MustParse("ny"):      {},
	language.MustParse("fj"):      {},
	language.MustParse("ee"):      {},
	language.MustParse("ak"):      {},
	language.MustParse("ks"):      {},
	language.MustParse("sg"):      {},
	language.MustParse("ik"):      {},
	language.MustParse("ve"):      {},
	language.MustParse("dz"):      {},
	language.MustParse("ff"):      {},
	language.MustParse("ti"):      {},
	language.MustParse("cr"):      {},
	language.MustParse("ng"):      {},
	language.MustParse("cho"):     {},
	language.MustParse("kj"):      {},
	language.MustParse("mh"):      {},
	language.MustParse("ho"):      {},
	language.MustParse("ii"):      {},
	language.MustParse("aa"):      {},
	language.MustParse("mus"):     {},
	language.MustParse("hz"):      {},
	language.MustParse("kr"):      {},
	language.MustParse("hil"):     {},
	language.MustParse("kbp"):     {},
	language.MustParse("din"):     {},
}
