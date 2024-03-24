package frontend

import "net/http"

func (f *Frontend) openSearchHandler(w http.ResponseWriter, r *http.Request) *response {
	resp := &response{
		status: http.StatusOK,
		data: data{
			Brand: f.Brand,
		},
		template: "opensearch",
		err:      nil,
	}

	return resp
}
