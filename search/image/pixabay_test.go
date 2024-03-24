package image

import (
	"net/http"
	"reflect"
	"testing"

	"github.com/jarcoal/httpmock"
)

func TestPixabayFetch(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	type args struct {
		query  string
		safe   bool
		number int
		offset int
	}

	for _, tt := range []struct {
		name string
		args
		u      string
		status int
		resp   string
		want   *Results
	}{
		{
			name:   "cat",
			args:   args{"cat", true, 100, 100},
			u:      `https://pixabay.com/api/?key=test&page=2&per_page=100&q=cat&safesearch=true`,
			status: 200,
			resp: `{"totalHits":500,"hits":[
					{"largeImageURL":"https://pixabay.com/get/ea33b90e28f5053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_1280.jpg","webformatHeight":426,"webformatWidth":640,"likes":42,"imageWidth":4896,"id":3681014,"user_id":1195798,"views":1570,"comments":21,"pageURL":"https://pixabay.com/photos/milk-can-old-pot-deformed-3681014/","imageHeight":3264,"webformatURL":"https://pixabay.com/get/ea33b90e28f5053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_640.jpg","type":"photo","previewHeight":99,"tags":"milk can, old, pot","downloads":826,"user":"Couleur","favorites":21,"imageSize":2862898,"previewWidth":150,"userImageURL":"https://cdn.pixabay.com/user/2019/02/12/21-34-01-586_250x250.jpg","previewURL":"https://cdn.pixabay.com/photo/2018/09/16/09/59/milk-can-3681014_150.jpg"},
					{"largeImageURL":"https://pixabay.com/get/e837b90e2ef7053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_1280.jpg","webformatHeight":398,"webformatWidth":640,"likes":27,"imageWidth":3119,"id":1281634,"user_id":2286921,"views":9643,"comments":0,"pageURL":"https://pixabay.com/photos/person-sport-bike-bicycle-cyclist-1281634/","imageHeight":1943,"webformatURL":"https://pixabay.com/get/e837b90e2ef7053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_640.jpg","type":"photo","previewHeight":93,"tags":"person, sport, bike","downloads":4105,"user":"Pexels","favorites":43,"imageSize":1400111,"previewWidth":150,"userImageURL":"https://cdn.pixabay.com/user/2016/03/26/22-06-36-459_250x250.jpg","previewURL":"https://cdn.pixabay.com/photo/2016/03/26/22/33/person-1281634_150.jpg"}
				],
				"total":4156
			}`,
			want: &Results{
				Provider: PixabayProvider,
				Count:    4156,
				Images: []*Image{
					{
						ID: "https://pixabay.com/get/ea33b90e28f5053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_640.jpg",
					},
					{
						ID: "https://pixabay.com/get/e837b90e2ef7053ed1584d05fb1d4797ea70e7d610b00c4090f5c27ea7e4b6bfda_640.jpg",
					},
				},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			responder := httpmock.NewStringResponder(tt.status, tt.resp)
			httpmock.RegisterResponder("GET", tt.u, responder)

			p := &Pixabay{
				Key:        "test",
				HTTPClient: &http.Client{},
			}
			got, err := p.Fetch(tt.args.query, tt.args.safe, tt.args.number, tt.args.offset)
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
