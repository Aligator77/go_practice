package models

import "encoding/json"

type Redirect struct {
	ID         string `json:"uuid"`
	IsDelete   int    `json:"is_deleted"`
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
