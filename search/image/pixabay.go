package image

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
)

// Pixabay holds settings for the Pixabay image search API
type Pixabay struct {
	Key        string
	HTTPClient *http.Client
}

// PixabayProvider is an image provider
const PixabayProvider Provider = "Pixabay"

// Fetch returns image results for a search query
func (p *Pixabay) Fetch(query string, safe bool, number int, offset int) (*Results, error) {
	u, err := url.Parse("https://pixabay.com/api/")
	if err != nil {
		return nil, err
	}

	var safeSearch = "true"
	if !safe {
		safeSearch = "false"
	}

	q := u.Query()
	q.Set("key", p.Key)
	q.Set("q", query)
	q.Set("per_page", strconv.Itoa(number))
	q.Set("page", strconv.Itoa((offset+number)/number))
	q.Set("safesearch", safeSearch)
	u.RawQuery = q.Encode()

	resp, err := p.HTTPClient.Get(u.String())
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bdy, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("Pixabay status: %d %q", resp.StatusCode, string(bdy))
	}

	var pr *PixabayResponse

	err = json.NewDecoder(resp.Body).Decode(&pr)
	if err != nil {
		return nil, err
	}

	res := &Results{
		Provider: PixabayProvider,
		Count:    pr.Total,
	}

	for _, h := range pr.Hits {
		img := &Image{
			ID: h.WebformatURL,
		}
		res.Images = append(res.Images, img)
	}

	return res, err
}

// PixabayResponse is the raw API response from Pixabay
type PixabayResponse struct {
	TotalHits int64 `json:"totalHits"`
	Hits      []struct {
		LargeImageURL   string `json:"largeImageURL"`
		WebformatHeight int    `json:"webformatHeight"`
		WebformatWidth  int    `json:"webformatWidth"`
		Likes           int    `json:"likes"`
		ImageWidth      int    `json:"imageWidth"`
		ID              int    `json:"id"`
		UserID          int    `json:"user_id"`
		Views           int    `json:"views"`
		Comments        int    `json:"comments"`
		PageURL         string `json:"pageURL"`
		ImageHeight     int    `json:"imageHeight"`
		WebformatURL    string `json:"webformatURL"`
		Type            string `json:"type"`
		PreviewHeight   int    `json:"previewHeight"`
		Tags            string `json:"tags"`
		Downloads       int64  `json:"downloads"`
		User            string `json:"user"`
		Favorites       int    `json:"favorites"`
		ImageSize       int64  `json:"imageSize"`
		PreviewWidth    int    `json:"previewWidth"`
		UserImageURL    string `json:"userImageURL"`
		PreviewURL      string `json:"previewURL"`
	} `json:"hits"`
	Total int64 `json:"total"`
}
