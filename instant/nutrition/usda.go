// Package nutrition provides food nutrition information
package nutrition

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/jivesearch/jivesearch/log"
)

// USDAProvider indicates the source is the USDA
const USDAProvider = "U.S. Department of Agriculture"

// USDA retrieves nutrition information from the USDA API
type USDA struct {
	Key        string
	HTTPClient *http.Client
}

// USDALookupResponse is the response from USDA
type USDALookupResponse struct {
	List struct {
		Q     string `json:"q"`
		Sr    string `json:"sr"`
		Ds    string `json:"ds"`
		Start int    `json:"start"`
		End   int    `json:"end"`
		Total int    `json:"total"`
		Group string `json:"group"`
		Sort  string `json:"sort"`
		Item  []struct {
			Offset       int    `json:"offset"`
			Group        string `json:"group"`
			Name         string `json:"name"`
			NDBNO        string `json:"ndbno"`
			Ds           string `json:"ds"`
			Manufacturer string `json:"manu"`
		} `json:"item"`
	} `json:"list"`
}

// USDAResponse is the response from USDA
type USDAResponse struct {
	Foods []struct {
		Food struct {
			Sr   string `json:"sr"`
			Type string `json:"type"`
			Desc struct {
				Ndbno string  `json:"ndbno"`
				Name  string  `json:"name"`
				Sd    string  `json:"sd"`
				Fg    string  `json:"fg"`
				Sn    string  `json:"sn"`
				Cn    string  `json:"cn"`
				Manu  string  `json:"manu"`
				Nf    float64 `json:"nf"`
				Cf    float64 `json:"cf"`
				Ff    float64 `json:"ff"`
				Pf    float64 `json:"pf"`
				R     string  `json:"r"`
				Rd    string  `json:"rd"`
				Ds    string  `json:"ds"`
				Ru    string  `json:"ru"`
			} `json:"desc"`
			Nutrients []struct {
				NutrientID json.Number   `json:"nutrient_id"` // sometimes it's a string, sometimes a number
				Name       string        `json:"name"`
				Group      string        `json:"group"`
				Unit       string        `json:"unit"`
				Value      json.Number   `json:"value"`
				Derivation string        `json:"derivation"`
				Sourcecode interface{}   `json:"sourcecode"`
				Dp         json.Number   `json:"dp"` // sometimes it's a string, sometimes a number
				Se         string        `json:"se"`
				Measures   []USDAMeasure `json:"measures"`
			} `json:"nutrients"`
			Sources []struct {
				ID      int    `json:"id"`
				Title   string `json:"title"`
				Authors string `json:"authors"`
				Vol     string `json:"vol"`
				Iss     string `json:"iss"`
				Year    string `json:"year"`
			} `json:"sources"`
			Footnotes []interface{} `json:"footnotes"`
			Langual   []interface{} `json:"langual"`
		} `json:"food"`
	} `json:"foods"`
	Count    int     `json:"count"`
	Notfound int     `json:"notfound"`
	API      float64 `json:"api"`
}

// USDAMeasure is an individual size
type USDAMeasure struct {
	Label string      `json:"label"`
	Eqv   float64     `json:"eqv"`
	Eunit string      `json:"eunit"`
	Qty   float64     `json:"qty"`
	Value json.Number `json:"value"`
}

// Can lookup 25 ndbno per request
const usdaMax = "25"

// Fetch retrieves nutrition information form USDA's API
func (u *USDA) Fetch(ndbnos []string) (*Response, error) {
	uu, err := url.Parse("https://api.nal.usda.gov/ndb/V2/reports")
	if err != nil {
		return nil, err
	}

	q := uu.Query()
	q.Set("api_key", u.Key)
	q.Set("max", usdaMax)
	q.Set("type", "f")
	q.Set("format", "json")
	uu.RawQuery = q.Encode()

	for _, n := range ndbnos {
		uu.RawQuery += "&ndbno=" + n
	}

	log.Debug.Println(uu.String())

	resp, err := u.HTTPClient.Get(uu.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	ur := &USDAResponse{}
	if err := json.Unmarshal(bdy, &ur); err != nil {
		return nil, err
	}

	r := &Response{
		Provider: USDAProvider,
	}

	for _, food := range ur.Foods {
		f := Food{
			Name:        food.Food.Desc.Name,
			FoodGroup:   food.Food.Desc.Fg,
			Corporation: food.Food.Desc.Manu,
		}

		for _, nutrient := range food.Food.Nutrients {
			n := Nutrient{
				ID:    nutrient.NutrientID,
				Name:  nutrient.Name,
				Unit:  nutrient.Unit,
				Value: nutrient.Value,
			}

			for _, measure := range nutrient.Measures {
				if (measure == USDAMeasure{}) {
					continue
				}
				m := Measure{
					Label:      measure.Label,
					Equivalent: measure.Eqv,
					Units:      measure.Eunit,
					Quantity:   measure.Qty,
					Value:      measure.Value,
				}
				n.Measures = append(n.Measures, m)
			}

			f.Nutrients = append(f.Nutrients, n)
		}

		r.Foods = append(r.Foods, f)
	}

	return r, err
}

// Lookup finds the USDA code(s) for an item
func (u *USDA) Lookup(query string) ([]*ItemResponse, error) {
	items := []*ItemResponse{}

	uu, err := url.Parse("https://api.nal.usda.gov/ndb/search/")
	if err != nil {
		return items, err
	}

	q := uu.Query()
	q.Set("api_key", u.Key)
	q.Set("q", query)
	q.Set("max", usdaMax)
	q.Set("sort", "n")
	q.Set("format", "json")
	q.Set("offset", "0")
	uu.RawQuery = q.Encode()

	log.Debug.Println(uu.String())

	resp, err := u.HTTPClient.Get(uu.String())
	if err != nil {
		return items, err
	}

	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return items, err
	}

	l := &USDALookupResponse{}

	if err := json.Unmarshal(bdy, &l); err != nil {
		return items, err
	}

	for _, j := range l.List.Item {
		item := &ItemResponse{
			Name:         j.Name,
			NDBNO:        j.NDBNO,
			Manufacturer: j.Manufacturer,
		}
		items = append(items, item)
	}

	return items, err
}
