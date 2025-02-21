package models

import "net/http"

type UrlData struct {
	Url interface{}
}

func (u UrlData) Bind(r *http.Request) error {
	//TODO implement me
	panic("implement me")
}
