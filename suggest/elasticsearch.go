package suggest

import (
	"context"
	"fmt"

	"github.com/olivere/elastic/v7"
)

const completionSuggest = "completion_suggest"

// ElasticSearch holds the index name and the connection
type ElasticSearch struct {
	Client *elastic.Client
	Index  string
	Type   string
}

// Completion handles autocomplete queries
func (e *ElasticSearch) Completion(term string, size int) (Results, error) {
	// Another option is the NewFuzzyCompletionSuggester and
	// set the "Fuzziness" but we'll start with this for now.
	res := Results{}

	s := elastic.NewCompletionSuggester(completionSuggest).
		Text(term).
		Field(completionSuggest).
		Size(size)

	result, err := e.Client.
		Search().
		Index(e.Index).
		Query(elastic.NewMatchAllQuery()).
		Suggester(s).
		Do(context.TODO())

	if err != nil {
		return res, err
	}

	if item, ok := result.Suggest[completionSuggest]; ok {
		for _, sug := range item {
			for _, opt := range sug.Options {
				res.Suggestions = append(res.Suggestions, opt.Text)
			}
		}
	}

	return res, nil
}

// Exists checks if a term is already in our index
func (e *ElasticSearch) Exists(term string) (bool, error) {
	return e.Client.Exists().
		Index(e.Index).
		Type(e.Type).
		Id(term).
		Do(context.TODO())
}

// Insert adds a new term to our index
func (e *ElasticSearch) Insert(term string) error {
	q := struct {
		Completion *elastic.SuggestField `json:"completion_suggest"`
	}{
		elastic.NewSuggestField().Input(term).Weight(0),
	}

	_, err := e.Client.Index().
		Index(e.Index).
		Type(e.Type).
		Id(term).
		BodyJson(&q).
		Do(context.TODO())

	return err
}

// Increment increments a term in our index
func (e *ElasticSearch) Increment(term string) error {
	_, err := e.Client.
		Update().
		Index(e.Index).
		Id(term).
		Script(elastic.NewScriptInline("ctx._source.completion_suggest.weight += 1")).
		Do(context.TODO())

	return err
}

func (e *ElasticSearch) mapping() string {
	mapping := fmt.Sprintf(`{
		"mappings": {
			"%v": {
				"dynamic": "strict",
				"properties": {
					"%v": {
						"type": "completion",
						"analyzer": "simple",
						"search_analyzer" : "simple",
						"preserve_separators": true,
						"preserve_position_increments": true,
						"max_input_length": 50
					}					
				}
			}
		}
	}`, e.Type, completionSuggest)

	fmt.Println(mapping)

	return mapping
}

// Setup creates a completion index
func (e *ElasticSearch) Setup() error {
	_, err := e.Client.CreateIndex(e.Index).Body(e.mapping()).Do(context.TODO())
	return err
}

// IndexExists returns true if the index exists
func (e *ElasticSearch) IndexExists() (bool, error) {
	return e.Client.IndexExists(e.Index).Do(context.TODO())
}
