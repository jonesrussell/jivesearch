package main

import (
	"context"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"
	"sync"

	"github.com/jivesearch/jivesearch/config"
	"github.com/jivesearch/jivesearch/log"
	img "github.com/jivesearch/jivesearch/search/image"
	"github.com/olivere/elastic"
	"github.com/spf13/viper"
)

type conf struct {
	e       *img.ElasticSearch
	workers int
	client  *http.Client
	host    string
	since   time.Time
	ch      chan *img.Image
}

var c *conf

func afterFn(executionID int64, requests []elastic.BulkableRequest, resp *elastic.BulkResponse, err error) {
	// NOTE: err can be nil even if documents fail to update
	if resp != nil {
		failed := resp.Failed()
		for _, d := range failed {
			log.Info.Printf("document failed: %+v\n", d)
			log.Info.Printf(" reason: %+v\n", d.Error)
		}
	}

	if err != nil {
		panic(err)
	}
}

func setup(v *viper.Viper) {
	v.SetEnvPrefix("jivesearch")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	config.SetDefaults(v)

	if v.GetBool("debug") {
		log.Debug.SetOutput(os.Stdout)
	}

	c = &conf{
		workers: v.GetInt("nsfw.workers"),
		client: &http.Client{
			Timeout: 25 * time.Second,
		},
		host:  v.GetString("nsfw.host"),
		since: v.GetTime("nsfw.since"),
		ch:    make(chan *img.Image),
	}
}

func main() {
	v := viper.New()
	setup(v)

	client, err := elastic.NewClient(elastic.SetURL(v.GetString("elasticsearch.url")), elastic.SetSniff(false))
	if err != nil {
		panic(err)
	}

	bulkActions := 1000
	bulk, err := client.BulkProcessor().
		After(afterFn).
		BulkActions(bulkActions).
		Do(context.Background())

	if err != nil {
		panic(err)
	}

	defer bulk.Close()

	c.e = &img.ElasticSearch{
		Client: client,
		Index:  v.GetString("elasticsearch.image.index"),
		Type:   v.GetString("elasticsearch.image.type"),
		Bulk:   bulk,
	}

	var wg sync.WaitGroup
	for worker := 0; worker < c.workers; worker++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()

			for lnk := range c.ch {
				im, err := c.fetchImage(lnk)
				if err != nil {
					log.Info.Println(err)
				}

				if err = c.e.Upsert(im); err != nil {
					panic(err)
				}
			}
		}(worker)
	}

	for {
		images, err := c.e.Uncrawled(10000, c.since)
		if err != nil {
			panic(err)
		}

		if len(images) < 100 { // otherwise will keep crawling same links even w/ a manual flush & refresh.
			break
		}

		for _, im := range images {
			log.Info.Println(im.ID)
			c.ch <- im
		}

		// flush & refresh so that we don't keep recrawling
		// images if len(images) is < bulk refresh rate.
		if err = c.e.Bulk.Flush(); err != nil {
			panic(err)
		}

		if _, err = c.e.Client.Flush().Index(c.e.Index).Do(context.Background()); err != nil {
			panic(err)
		}

		if _, err = c.e.Client.Refresh().Index(c.e.Index).Do(context.Background()); err != nil {
			panic(err)
		}

		// not sure how to use the "wait_for" for indices refresh
		time.Sleep(2 * time.Second)
	}

	close(c.ch)
	wg.Wait()
}

func (c *conf) fetchImage(i *img.Image) (*img.Image, error) {
	i.Crawled = time.Now().Format("20060102")

	// The following is a pure Go attempt to cut out the Python server altogether
	// but doesn't seem to work very well:
	// https://gist.github.com/brentadamson/e601d8a2704e10dcd2d3ea5c31301994
	u := c.host + "?image=" + i.ID
	resp, err := c.client.Get(u)
	if err != nil {
		return i, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		bdy, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return i, err
		}

		im := &img.Image{}

		if err := json.Unmarshal(bdy, &im); err != nil {
			return i, err
		}

		i.NSFW = math.Round(im.NSFW/.0001) / 10000
		i.Width = im.Width
		i.Height = im.Height
		i.MIME = im.MIME
		i.EXIF = im.EXIF
		i.Classification = separateKeys(im.Classification)
	}

	return i, err
}

// separateKeys turns "punching bag, punch bag" to 2 items
// In case of duplicate keys we take that with highest value
func separateKeys(c map[string]float64) map[string]float64 {
	m := map[string]float64{}
	for key, val := range c {
		rounded := math.Round(val/.001) / 1000
		for _, s := range strings.Split(key, ",") {
			s = strings.TrimSpace(s)
			if v, ok := m[s]; ok {
				if v >= rounded {
					continue
				}
			}
			m[s] = rounded
		}
	}

	return m
}
