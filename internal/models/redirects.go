package models

import "encoding/json"

type Redirect struct {
	ID         string `json:"uuid"`
	IsActive   int    `json:"is_active"`
	URL        string `json:"url"`
	Redirect   string `json:"redirect"`
	DateCreate string `json:"dateCreate"`
	DateUpdate string `json:"dateUpdate"`
}

func (r Redirect) String() string {
	res, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(res)
}
