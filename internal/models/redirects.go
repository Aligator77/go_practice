// Package models contain models for all project
package models

import "encoding/json"

type Redirect struct {
	ID         string `json:"uuid"`
	IsDelete   int    `json:"is_deleted"` // change for iter15
	URL        string `json:"url"`
	Redirect   string `json:"redirect"`
	DateCreate string `json:"dateCreate"`
	DateUpdate string `json:"dateUpdate"`
	User       string `json:"user"`
}

func (r Redirect) String() string {
	res, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(res)
}
