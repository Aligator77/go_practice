package models

type Redirect struct {
	IsActive   int    `json:"is_active"`
	Url        string `json:"url"`
	Redirect   string `json:"redirect"`
	DateCreate string `json:"dateCreate"`
	DateUpdate string `json:"dateUpdate"`
}
