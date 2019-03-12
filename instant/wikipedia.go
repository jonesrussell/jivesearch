package instant

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	timezone "github.com/evanoberholster/timezoneLookup"
	"github.com/jivesearch/jivesearch/instant/location"
	"github.com/jivesearch/jivesearch/instant/nutrition"
	"github.com/jivesearch/jivesearch/instant/wikipedia"
	"golang.org/x/text/language"
)

// WikipediaType is a Wikipedia answer Type
const (
	WikipediaType         Type = "wikipedia"
	WikidataAgeType       Type = "wikidata age"
	WikidataBirthdayType  Type = "wikidata birthday"
	WikidataClockType     Type = "wikidata clock"
	WikidataDeathType     Type = "wikidata death"
	WikidataHeightType    Type = "wikidata height"
	WikidataNutritionType Type = "wikidata nutrition"
	WikidataWeightType    Type = "wikidata weight"
	WikiquoteType         Type = "wikiquote"
	WiktionaryType        Type = "wiktionary"
)

// Wikipedia is a Wiki* instant answer,
// including Wikidata/Wikiquote/Wiktionary
type Wikipedia struct {
	LocationFetcher  location.Fetcher
	NutritionFetcher nutrition.Fetcher
	TimeZoneFetcher  timezone.TimezoneInterface
	wikipedia.Fetcher
	Answer
}

func (w *Wikipedia) setQuery(r *http.Request, qv string) Answerer {
	w.Answer.setQuery(r, qv)
	return w
}

func (w *Wikipedia) setUserAgent(r *http.Request) Answerer {
	return w
}

func (w *Wikipedia) setLanguage(lang language.Tag) Answerer {
	w.language = lang
	return w
}

func (w *Wikipedia) setType() Answerer {
	w.Type = WikipediaType
	return w
}

// trigger words
// age ---> for "how old is x?" we need to change our triggerfuncs to just a regex
const age = "age"
const howOldIs = "how old is"

// birthday
const birthday = "birthday"
const born = "born"

// clock
const clock = "clock"
const currentTime = "current time"
const timeIn = "time in"
const wTime = "time"

// death
const death = "death"
const died = "died"

// height
const height = "height"
const howTallis = "how tall is"
const howTallwas = "how tall was"

// Nutrient is an individual nutrient
type Nutrient struct {
	Name string
	Code string
}

// Nutrients is all the nutrients
type Nutrients []Nutrient

const (
	calcium       = "calcium"
	calories      = "calories"
	carbs         = "carbs"
	carbohydrates = "carbohydrates"
	cholesterol   = "cholesterol"
	saturatedFat  = "saturated fat"
	fat           = "fat"
	fiber         = "fiber"
	iron          = "iron"
	lipid         = "lipid"
	magnesium     = "magnesium"
	potassium     = "potassium"
	protein       = "protein"
	sodium        = "sodium"
	sugars        = "sugars"
	sugar         = "sugar"
	vitaminA      = "vitamin a"
	vitaminB      = "vitamin b"
	vitaminB12    = "vitamin b12"
	vitaminBB12   = "vitamin b-12"
	vitaminBBB12  = "vitamin b 12"
	vitaminC      = "vitamin c"
	vitaminD      = "vitamin d"
	zinc          = "zinc"
)

var nutritionTriggers = []string{
	calcium, calories, carbs, carbohydrates, cholesterol, saturatedFat, fat, fiber, iron, lipid,
	magnesium, potassium, protein, sodium, sugars, sugar, vitaminA,
	vitaminB, vitaminB12, vitaminBB12, vitaminBBB12, vitaminC, vitaminD, zinc,
}

// weight
// will fail on "how much does x weigh?"
const mass = "mass"
const weigh = "weigh"
const weight = "weight"

// quotes
const quote = "quote"
const quotes = "quotes"

// definitions
const define = "define"
const definition = "definition"

func (w *Wikipedia) setRegex() Answerer {
	triggers := []string{
		age, howOldIs,
		birthday, born,
		death, died,
		howTallis, howTallwas, height,
		mass, weigh, weight,
		clock, currentTime, timeIn, wTime,
		quote, quotes,
		define, definition,
	}

	t := strings.Join(triggers, "|")

	for _, t := range nutritionTriggers {
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^how (many|much) (?P<trigger>%s) are in a (?P<remainder>.*)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^how (many|much) (?P<trigger>%s) in a (?P<remainder>.*)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^how (many|much) (?P<trigger>%s) in (?P<remainder>.*)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) in a (?P<remainder>.*)$`, t)))
		w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) in (?P<remainder>.*)$`, t)))
	}

	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<trigger>%s) (?P<remainder>.*)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(fmt.Sprintf(`^(?P<remainder>.*) (?P<trigger>%s)$`, t)))
	w.regex = append(w.regex, regexp.MustCompile(`^(?P<remainder>.*)$`)) // this needs to be last regex here

	return w
}

// Age is a person's current age (in years) or age when they died
type Age struct {
	*Birthday `json:"birthday,omitempty"`
	*Death    `json:"death,omitempty"`
}

// Birthday is a person's date of birth
type Birthday struct {
	Birthday wikipedia.DateTime `json:"birthday,omitempty"`
}

// Clock is a current time for a location
// TODO: make Location a map/struct of different languages, not just 1
type Clock struct {
	Time     time.Time `json:"time"`
	Location struct {
		City    string `json:"city"`
		State   string `json:"state"`
		Country string `json:"country"`
	}
}

// Death is a person's date of death
// TODO: add place of death, cause, etc.
type Death struct {
	Death wikipedia.DateTime `json:"death,omitempty"`
}

var contains = func(x string, sl []string) bool {
	for _, s := range sl {
		if x == s {
			return true
		}
	}
	return false
}

// TODO: Return the Title (and perhaps Image???) as
// confirmation that we fetched the right asset.
func (w *Wikipedia) solve(r *http.Request) Answerer {
	items, err := w.Fetch(w.remainder, w.language)
	if err != nil {
		w.Err = err
		return w
	}

	switch w.triggerWord {
	case age, howOldIs, birthday, born:
		b := &Birthday{}

		for _, item := range items {
			if len(item.Birthday) == 0 {
				return w
			}
			w.Type = WikidataBirthdayType
			b.Birthday = item.Birthday[0]
		}

		if w.triggerWord == "age" || w.triggerWord == "how old is" {
			w.Type = WikidataAgeType

			a := &Age{
				Birthday: b,
			}

			for _, item := range items {
				if len(item.Death) > 0 {
					a.Death = &Death{item.Death[0]}
				}
			}

			w.Data.Solution = a

			return w
		}

		w.Data.Solution = b
	case clock, currentTime, timeIn, wTime:
		var err error
		cc := &Clock{}

		var lat, lon float32

		for _, item := range items {
			for _, c := range item.Instance {
				switch c.ID {
				case "Q515", "Q1093829": // a city
					cc.Location.City = item.Labels["en"].Text
					for _, c := range item.Country {
						for _, item := range c.Item {
							cc.Location.Country = string(item.Labels["en"].Text)
						}
					}
				case "Q6256": // a country
					cc.Location.Country = item.Labels["en"].Text
					for _, c := range item.Capital {
						// get the current capital city...e.g. has no end date
						if len(c.End) == 0 {
							for _, cap := range c.Item {
								cc.Location.City = cap.Labels["en"].Text
							}
						}
					}
				}

			}
			for _, c := range item.Coordinate {
				lat = float32(c.Latitude[0])
				lon = float32(c.Longitude[0])
			}

		}

		cc.Time, err = w.getTime(lat, lon)
		if err != nil {
			w.Err = err
			return w
		}

		w.Type = WikidataClockType
		w.Data.Solution = cc
	case death, died:
		for _, item := range items {
			if len(item.Death) > 0 {
				w.Type = WikidataDeathType
				w.Data.Solution = &Death{item.Death[0]}
			}
		}
	case howTallis, howTallwas, height:
		for _, item := range items {
			if len(item.Height) == 0 {
				return w
			}
			w.Type = WikidataHeightType
			w.Data.Solution = item.Height
		}
	case calcium, calories, carbs, carbohydrates, cholesterol, saturatedFat, fat, fiber, iron, lipid,
		magnesium, potassium, protein, sodium, sugars, sugar, vitaminA,
		vitaminB, vitaminB12, vitaminBB12, vitaminBBB12, vitaminC, vitaminD, zinc:
		// Wikipedia seems more reliable ndbno id's for items like "egg"
		// but doesn't have things like "Whopper" or "Big Mac without sauce"
		var ndbnos = []string{}
		switch w.remainder {
		case "cheese":

		case "egg", "eggs":
			ndbnos = []string{"01123"}
		default:
			var empty = true
			for _, item := range items {
				for _, u := range item.USDA {
					ndbnos = append(ndbnos, u)
					empty = false
				}
			}

			// if it is a branded product, then get the rest of that brand
			// e.g. Show McDonald's Big Mac w/out sauce
			itms, err := w.NutritionFetcher.Lookup(w.remainder)
			if err != nil {
				w.Err = err
				return w
			}

			var manufacturer string

			for _, ndbno := range ndbnos {
				for _, itm := range itms {
					if itm.NDBNO == ndbno {
						manufacturer = itm.Manufacturer
					}
				}
			}

			for _, itm := range itms {
				if contains(itm.NDBNO, ndbnos) {
					continue
				}

				switch empty {
				case true:
					ndbnos = append(ndbnos, itm.NDBNO)
				default:
					if manufacturer != "" && itm.Manufacturer == manufacturer {
						ndbnos = append(ndbnos, itm.NDBNO)
					}
				}
			}

		}

		if len(ndbnos) == 0 {
			w.Err = fmt.Errorf("unable to find ndbno identifier for %v", w.remainder)
			return w
		}

		resp, err := w.NutritionFetcher.Fetch(ndbnos)
		if err != nil {
			w.Err = err
			return w
		}

		resp.Trigger = w.triggerWord
		w.Type = WikidataNutritionType
		w.Data.Solution = resp
	case mass, weigh, weight:
		for _, item := range items {
			if len(item.Weight) == 0 {
				return w
			}
			w.Type = WikidataWeightType
			w.Data.Solution = item.Weight
		}
	case quote, quotes:
		for _, item := range items {
			if len(item.Wikiquote.Quotes) == 0 {
				return w
			}
			w.Type = WikiquoteType
			w.Data.Solution = item.Wikiquote.Quotes
		}
	case define, definition:
		for _, item := range items {
			if len(item.Wiktionary.Definitions) == 0 {
				return w
			}
			w.Type = WiktionaryType
			w.Data.Solution = item.Wiktionary
		}
	default: // full Wikipedia box unless for certain words
		switch w.remainder {
		case clock, currentTime, timeIn, wTime: // if "clock", "current time", etc. then they want the current time for their current location
			ip := getIPAddress(r)
			c, err := w.LocationFetcher.Fetch(ip)
			if err != nil {
				w.Err = err
				return w
			}

			lat, lon := float32(c.Location.Latitude), float32(c.Location.Longitude)
			t, err := w.getTime(lat, lon)
			if err != nil {
				w.Err = err
				return w
			}

			w.Type = WikidataClockType
			cc := &Clock{
				Time: t,
			}

			lang := "en"
			cc.Location.City = c.City.Names[lang]
			for _, s := range c.Subdivisions {
				cc.Location.State = s.Names[lang]
			}

			cc.Location.Country = c.Country.Names[lang]

			w.Data.Solution = cc
		default:
			w.Type = WikipediaType
			w.Data.Solution = items
		}
	}

	return w
}

func (w *Wikipedia) getTime(lat, lon float32) (time.Time, error) {
	t := time.Time{}

	zone, err := w.TimeZoneFetcher.Query(
		timezone.Coord{Lat: lon, Lon: lat}, // is this backwards or just me???
	)

	if err != nil {
		return t, err
	}

	loc, err := time.LoadLocation(zone)
	if err != nil {
		return t, err
	}

	t = now().In(loc)
	return t, nil
}

func (w *Wikipedia) tests() []test {
	sydney, err := time.LoadLocation("Australia/Sydney")
	if err != nil {
		panic(err)
	}

	mountain, err := time.LoadLocation("America/Denver")
	if err != nil {
		panic(err)
	}

	timeInUTC := time.Date(2016, 6, 5, 3, 2, 0, 0, time.UTC)

	cl := &Clock{
		Time: timeInUTC.In(mountain),
	}
	cl.Location.City = "Someville"
	cl.Location.State = "SomeState"
	cl.Location.Country = "SomeCountry"

	tests := []test{
		{
			query: "Bob Marley age",
			expected: []Data{
				{
					Type:      WikidataAgeType,
					Triggered: true,
					Solution: &Age{
						Birthday: &Birthday{
							Birthday: wikipedia.DateTime{
								Value:    "1945-02-06T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
						Death: &Death{
							Death: wikipedia.DateTime{
								Value:    "1981-05-11T00:00:00Z",
								Calendar: wikipedia.Wikidata{ID: "Q1985727"},
							},
						},
					},
				},
			},
		},
		{
			query: "Jimi hendrix birthday",
			expected: []Data{
				{
					Type:      WikidataBirthdayType,
					Triggered: true,
					Solution: &Birthday{
						Birthday: wikipedia.DateTime{
							Value:    "1942-11-27T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		},
		{
			query: "death jimi hendrix",
			expected: []Data{
				{
					Type:      WikidataDeathType,
					Triggered: true,
					Solution: &Death{
						Death: wikipedia.DateTime{
							Value:    "1970-09-18T00:00:00Z",
							Calendar: wikipedia.Wikidata{ID: "Q1985727"},
						},
					},
				},
			},
		},
		{
			query: "shaquille o'neal height",
			expected: []Data{
				{
					Type:      WikidataHeightType,
					Triggered: true,
					Solution: []wikipedia.Quantity{
						{
							Amount: "2.16",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
				},
			},
		},
		{
			query: "shaquille o'neal weight",
			expected: []Data{
				{
					Type:      WikidataWeightType,
					Triggered: true,
					Solution: []wikipedia.Quantity{
						{
							Amount: "147",
							Unit:   wikipedia.Wikidata{ID: "Q11573"},
						},
					},
				},
			},
		},
		{
			query: "Michael Jordan quotes",
			expected: []Data{
				{
					Type:      WikiquoteType,
					Triggered: true,
					Solution: []string{
						"I can accept failure. Everyone fails at something. But I can't accept not trying (no hard work)",
						"ball is life",
					},
				},
			},
		},
		{
			query: "define guitar",
			expected: []Data{
				{
					Type:      WiktionaryType,
					Triggered: true,
					Solution: wikipedia.Wiktionary{
						Title: "guitar",
						Definitions: []*wikipedia.Definition{
							{Part: "noun", Meaning: "musical instrument"},
						},
					},
				},
			},
		},
		{
			query: "jimi hendrix",
			expected: []Data{
				{
					Type:      WikipediaType,
					Triggered: true,
					Solution: []*wikipedia.Item{
						{
							Wikidata: &wikipedia.Wikidata{
								Claims: &wikipedia.Claims{
									Birthday: []wikipedia.DateTime{
										{
											Value:    "1942-11-27T00:00:00Z",
											Calendar: wikipedia.Wikidata{ID: "Q1985727"},
										},
									},
									Death: []wikipedia.DateTime{
										{
											Value:    "1970-09-18T00:00:00Z",
											Calendar: wikipedia.Wikidata{ID: "Q1985727"},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			query: "Sydney time",
			expected: []Data{
				{
					Type:      WikidataClockType,
					Triggered: true,
					Solution: &Clock{
						Time: timeInUTC.In(sydney),
					},
				},
			},
		},
		{
			query: "time",
			expected: []Data{
				{
					Type:      WikidataClockType,
					Triggered: true,
					Solution:  cl,
				},
			},
		},
		{
			query: "eggs sodium",
			expected: []Data{
				{
					Type:      WikidataNutritionType,
					Triggered: true,
					Solution: &nutrition.Response{
						Trigger: "sodium",
						Foods: []nutrition.Food{
							{
								Name:        "Egg, whole, raw, fresh",
								FoodGroup:   "Dairy and Egg Products",
								Corporation: "",
								Nutrients: []nutrition.Nutrient{
									{
										ID:    "212",
										Name:  "Sodium",
										Unit:  "mg",
										Value: json.Number("12"),
										Measures: []nutrition.Measure{
											{
												Label:      "large",
												Equivalent: 50,
												Units:      "g",
												Quantity:   1,
												Value:      json.Number("72"),
											},
											{
												Label:      "extra large",
												Equivalent: 56,
												Units:      "g",
												Quantity:   1,
												Value:      json.Number("80"),
											},
										},
									},
								},
							},
						},
						Provider: "Mock Response",
					},
				},
			},
		},
		{
			query: "big mac calories",
			expected: []Data{
				{
					Type:      WikidataNutritionType,
					Triggered: true,
					Solution: &nutrition.Response{
						Trigger: "calories",
						Foods: []nutrition.Food{
							{
								Name:        "Big Mac",
								FoodGroup:   "Some Category",
								Corporation: "McDowell's",
								Nutrients: []nutrition.Nutrient{
									{
										ID:    "208",
										Name:  "Energy",
										Unit:  "kcal",
										Value: json.Number("554"),
										Measures: []nutrition.Measure{
											{
												Label:      "1 size",
												Equivalent: 12,
												Units:      "g",
												Quantity:   1,
												Value:      json.Number("720"),
											},
										},
									},
								},
							},
						},
						Provider: "Mock Response",
					},
				},
			},
		},
	}

	return tests
}
