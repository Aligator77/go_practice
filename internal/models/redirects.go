package models

type Redirect struct {
	Url        string `json:"url"`
	Redirect   string `json:"redirect"`
	DateCreate string `json:"dateCreate"`
	DateUpdate string `json:"dateUpdate"`
}
