package bangs

import (
	"context"
	"fmt"
	"strings"

	"github.com/olivere/elastic"
)

// Elasticsearch is probably overkill for this but oh, well.

const bangSuggest = "bang_suggest"

// ElasticSearch holds the index name and the connection
type ElasticSearch struct {
	Client *elastic.Client
	Index  string
	Type   string
}

// SuggestResults retrieves !bang suggestions from Elasticsearch
func (e *ElasticSearch) SuggestResults(term string, size int) (Results, error) {
	// Another option is the NewFuzzyCompletionSuggester and
	// set the "Fuzziness" but we'll start with this for now.
	res := Results{}

	term = strings.TrimPrefix(term, "!")

	s := elastic.NewCompletionSuggester(bangSuggest).
		Text(term).
		Field(bangSuggest).
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

	if item, ok := result.Suggest[bangSuggest]; ok {
		for _, sug := range item {
			for _, opt := range sug.Options {
				sug := Suggestion{
					Trigger: opt.Text,
				}

				res.Suggestions = append(res.Suggestions, sug)
			}
		}
	}

	return res, nil
}

func (e *ElasticSearch) mapping() string {
	// Note: the simple analyzer won't work for numeric triggers -> e.g. "500".
	return fmt.Sprintf(`{
		"mappings": {
			"%v": {
				"dynamic": "strict",
				"properties": {
					"%v": {
						"type": "completion",
						"analyzer": "standard",
						"search_analyzer" : "standard",
						"preserve_separators": true,
						"preserve_position_increments": true,
						"max_input_length": 100
					}
				}
			}
		}
	}`, e.Type, bangSuggest)
}

// IndexExists returns true if the index exists
func (e *ElasticSearch) IndexExists() (bool, error) {
	return e.Client.IndexExists(e.Index).Do(context.TODO())
}

// DeleteIndex will delete the existing index
func (e *ElasticSearch) DeleteIndex() error {
	_, err := e.Client.DeleteIndex(e.Index).Do(context.TODO())
	return err
}

// Setup recreates the completion index
func (e *ElasticSearch) Setup(bangs []Bang) error {
	if _, err := e.Client.CreateIndex(e.Index).Body(e.mapping()).Do(context.TODO()); err != nil {
		return err
	}

	for _, b := range bangs {
		for _, t := range b.Triggers {
			q := struct {
				Completion *elastic.SuggestField `json:"bang_suggest"`
			}{
				elastic.NewSuggestField().Input(t).Weight(0),
			}

			_, err := e.Client.Index().
				Index(e.Index).
				Type(e.Type).
				BodyJson(&q).
				Do(context.TODO())

			if err != nil {
				return err
			}
		}
	}

	return nil
}
