package models

import (
	"io"
	"net/http"
)

type URLData struct {
	URL string `json:"url"`
}

func (u URLData) Bind(r *http.Request) error {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	//lint:ignore need for chi.Bind
	u.URL = string(url)

	return nil
}
