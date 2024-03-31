package image

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jivesearch/jivesearch/log"
	"github.com/olivere/elastic/v7"
)

// ElasticSearch hold connection and index settings
type ElasticSearch struct {
	Client        *elastic.Client
	Index         string
	Type          string
	Bulk          *elastic.BulkProcessor
	NSFWThreshold float64
}

// Fetch returns image results for a search query
func (e *ElasticSearch) Fetch(q string, safe bool, number int, offset int) (*Results, error) {
	res := &Results{}

	var safeQuery string

	switch safe {
	case false:
		safeQuery = fmt.Sprintf(`"gte": %v`, e.NSFWThreshold)
	default:
		safeQuery = fmt.Sprintf(`"lt": %v`, e.NSFWThreshold)
	}

	qu := fmt.Sprintf(`{
		"query": {
			"function_score": {
				"query": {
					"bool": {
						"should": [
							{
								"multi_match": {
									"query": "%v",
									"fields": [
										"alt"
									]
								}
							}
						],
						"must": [
							{
								"range": {
									"nsfw_score": {
										%v
									}
								}
							}
						]
					}
				},
				"field_value_factor": {
					"field": "classification.%v",
					"modifier": "log1p",
					"missing": 0.0,
					"factor": 2
				},
				"boost_mode": "sum"
			}
		},
		"from": %d,
		"size": %d
	}`, q, safeQuery, q, offset, number)

	out, err := e.Client.Search(e.Index).Source(qu).Do(context.TODO())
	if err != nil {
		return res, err
	}

	res.Count = out.TotalHits()

	for _, h := range out.Hits.Hits {
		img := &Image{
			ID: h.Id,
		}
		err := json.Unmarshal(h.Source, img)
		if err != nil {
			return res, err
		}

		res.Images = append(res.Images, img)
	}

	return res, err
}

// Upsert updates an image link or inserts it if it doesn't exist
// NOTE: Elasticsearch has a 512-byte limit on an insert operation.
// Upsert does not have that limit.
func (e *ElasticSearch) Upsert(img *Image) error {
	item := elastic.NewBulkUpdateRequest().
		Index(e.Index).
		Id(img.ID).
		DocAsUpsert(true).
		Doc(img)

	e.Bulk.Add(item)
	return nil
}

// Uncrawled finds images that haven't been crawled recently/yet
// We also aggregate by domain, which is equivalent to
// selecting unique domains so that we don't overload a domain.
// The subaggregation will then return 1 result for each domain.
func (e *ElasticSearch) Uncrawled(number int, since time.Time) ([]*Image, error) {
	agg := "by_domain"
	subAgg := "get_one"

	q := fmt.Sprintf(`{
		"query": {
		  "bool": {
			"should": [
			  {
				"bool": {
				  "must_not": [
					{
					  "exists": {
						"field": "crawled"
					  }
					}
				  ]
				}
			  },
			  {
				"bool": {
				  "filter": [
					{
					  "range": {
						"crawled": {
						  "lte": %v
						}
					  }
					}
				  ]
				}
			  }
			]
		  }
		},
		"aggs": {
		  "%v": {
				"terms": {
					"field": "domain",
					"size": %v
				},
				"aggs": {
					"%v": {
						"top_hits": {
							"_source": {
							"includes": [
								"id", "domain", "alt"
							]
							},
							"size": 1
						}
					}
				}
		  }
		},
		"size": 0
	}`, since.Format("20060102"), agg, number, subAgg)

	images := []*Image{}

	res, err := e.Client.Search(e.Index).Source(q).Do(context.TODO())
	if err != nil {
		return images, err
	}

	termsAggRes, found := res.Aggregations.Terms(agg)
	if !found || termsAggRes == nil {
		return images, fmt.Errorf("aggregation key not found")
	}

	for _, b := range termsAggRes.Buckets {
		hits, ok := b.TopHits(subAgg)
		if !ok {
			return images, fmt.Errorf("subaggregation key not found")
		}

		for _, h := range hits.Hits.Hits {
			img := &Image{}
			err := json.Unmarshal(h.Source, img)
			if err != nil {
				return images, err
			}

			img.ID = h.Id
			images = append(images, img)
		}
	}

	return images, nil
}

// Setup will create our image index
func (e *ElasticSearch) Setup() error {
	exists, err := e.Client.IndexExists(e.Index).Do(context.TODO())
	if err != nil {
		return err
	}

	if !exists {
		log.Info.Println("Creating index:", e.Index)
		if _, err = e.Client.CreateIndex(e.Index).Body(e.mapping()).Do(context.TODO()); err != nil {
			return err
		}
	}

	return nil
}

// mapping is the mapping of our image Index.
func (e *ElasticSearch) mapping() string {
	m := `{
			"mappings": {
					"properties": {
							"expires": {
									"type": "date",
									"format": "yyyyMMddHHmm"
							},
							"body": {
									"type": "text",
									"index": false
							},
							"status": {
									"type": "short"
							}
					}
			}
	}`

	fmt.Println(m)

	return m
}
