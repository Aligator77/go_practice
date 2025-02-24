package models

import (
	"io"
	"net/http"
)

type UrlData struct {
	Url string
}

func (u UrlData) Bind(r *http.Request) error {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	u.Url = string(url)

	return nil
}
