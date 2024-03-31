package crawler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/jonesrussell/jivesearch/search/document"
	"github.com/olivere/elastic/v7"
)

type ElasticSearch struct {
	*document.ElasticSearch
	Bulk *elastic.BulkProcessor
	sync.Mutex
}

func (e *ElasticSearch) Upsert(doc *document.Document) error {
	a, err := e.Analyzer(doc.Language)
	if err != nil {
		return err
	}

	idx := e.IndexName(a)

	item := elastic.NewBulkUpdateRequest().
		Index(idx).
		Type(e.Type).
		Id(doc.ID).
		DocAsUpsert(true).
		Doc(doc)

	e.Bulk.Add(item)
	return nil
}

func (e *ElasticSearch) CrawledAndCount(url, domain string) (time.Time, int, error) {
	if e == nil || e.Client == nil {
		return time.Time{}, 0, errors.New("ElasticSearch client is not initialized in CrawledAndCount function")
	}

	fmt.Println(domain)

	body := fmt.Sprintf(`{
		"bool": {
			"filter": [
				{
					"term": {
						"domain": "%v"
					}
				},
				{
					"term": {
						"index": "true"
					}
				}
			]
		}
	}`, domain)

	var crawled, cnt = time.Time{}, 0

	countReq := elastic.NewSearchRequest().
		Index(e.Index + "-*").
		Source(elastic.NewSearchSource().
			Query(elastic.RawStringQuery(body)),
		)

	crawledRequest := elastic.NewSearchRequest().
		Index(e.Index + "-*").
		Source(elastic.NewSearchSource().
			Query(elastic.NewTermQuery("_id", url)).
			FetchSourceContext(elastic.NewFetchSourceContext(true).Include("crawled")),
		)

	e.Lock()
	defer e.Unlock() // Ensure the lock is always released

	res, err := e.Client.MultiSearch().
		Add(countReq, crawledRequest).
		Do(context.TODO())

	if err != nil {
		fmt.Printf("Error executing multi-search: %v\n", err)
		return crawled, cnt, err
	}

	for i, r := range res.Responses {
		fmt.Printf("Response %d:\n", i+1)
		if r.Error != nil {
			fmt.Printf(" Error: %s\n", r.Error.Reason)
		} else {
			fmt.Printf(" Total Hits: %d\n", r.TotalHits())
			for _, hit := range r.Hits.Hits {
				fmt.Printf("    Hit ID: %s, Source: %s\n", hit.Id, string(hit.Source))
			}
		}
	}

	r1, r2 := res.Responses[0], res.Responses[1]

	cnt = int(r1.TotalHits())

	if !elastic.IsNotFound(r2.Error) {
		return crawled, cnt, fmt.Errorf("%v", r2.Error)
	}

	h2 := r2.Hits
	fmt.Printf("r2: %+v\n", h2)

	hits := r2.Hits.Hits

	for _, hit := range hits {
		fmt.Printf("h: %+v\n", hit)
		if hit.Source == nil {
			return crawled, cnt, fmt.Errorf("source is nil for hit ID: %s", hit.Id)
		}
		c := make(map[string]string)
		if err := json.Unmarshal(hit.Source, &c); err != nil {
			return crawled, cnt, err
		}
		crawledStr, ok := c["crawled"]
		if !ok {
			return crawled, cnt, fmt.Errorf("crawled field not found in hit ID: %s", hit.Id)
		}
		crawled, err = time.Parse("20060102", crawledStr)
		if err != nil {
			return crawled, cnt, err
		}

	}

	return crawled, cnt, err
}
