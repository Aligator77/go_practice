package models

import (
	"encoding/json"
	"io"
	"net/http"
)

type URLData struct {
	URL string `json:"url"`
}

type URLBatchData struct {
	CorrelationId string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}
type URLBatches struct {
	URLBatchData []URLBatchData
}

type URLDataResponse struct {
	Result string `json:"result"`
}

type URLBatchResponse struct {
	CorrelationId string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (u URLData) Bind(r *http.Request) error {
	url, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	//lint:ignore SA4000 need for chi.Bind
	u.URL = string(url)
	_ = u.URL

	return nil
}

func (u URLBatches) Bind(r *http.Request) error {
	links, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	var linksSlice []URLBatchData
	err = json.Unmarshal(links, &linksSlice)
	if err != nil {
		return err
	}

	return nil
}
