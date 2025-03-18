// Package models contain models for all project
package models

import (
	"io"
	"net/http"
)

type URLData struct {
	URL string `json:"url"`
}

type URLBatchData []struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type URLDataResponse struct {
	Result string `json:"result"`
}

type URLBatchResponse struct {
	CorrelationID string `json:"correlation_id,omitempty"`
	ShortURL      string `json:"short_url,omitempty"`
	OriginalURL   string `json:"original_url,omitempty"`
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
