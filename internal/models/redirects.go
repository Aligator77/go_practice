package models

type Redirect struct {
	ID         string `json:"uuid"`
	IsActive   int    `json:"is_active"`
	URL        string `json:"url"`
	Redirect   string `json:"redirect"`
	DateCreate string `json:"dateCreate"`
	DateUpdate string `json:"dateUpdate"`
}
