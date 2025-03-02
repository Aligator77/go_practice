package controllers

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"io"
	"net/http"
	"time"

	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/models"
	"github.com/Aligator77/go_practice/internal/server"
	"github.com/Aligator77/go_practice/internal/stores"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type URLController struct {
	URLService *stores.URLService
}

func NewURLController(URLService *stores.URLService) *URLController {
	return &URLController{
		URLService: URLService,
	}
}

func (u *URLController) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	newRedirect := helpers.GenerateRandomURL(10)
	redirect := &models.Redirect{
		IsActive:   1,
		URL:        string(data),
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	existRedirect, _ := u.URLService.GetRedirectByURL(redirect.URL)
	if len(existRedirect.URL) > 0 {
		render.Status(r, http.StatusConflict)
		w.WriteHeader(http.StatusConflict)

		_, err = w.Write([]byte(u.URLService.MakeFullURL(existRedirect.Redirect)))
		if err != nil {
			u.URLService.Logger.Err(err).Msg("Write error CreatePostHandler")
		}

		return
	}
	_, err = u.URLService.NewRedirect(*redirect)
	if err != nil {
		return
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(u.URLService.MakeFullURL(newRedirect)))
	if err != nil {
		return
	}
}

func (u *URLController) CreateRestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := &models.URLData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}

	newUUID, _ := uuid.NewV7()
	newRedirect := helpers.GenerateRandomURL(10)
	redirect := &models.Redirect{
		ID:         newUUID.String(),
		IsActive:   1,
		URL:        data.URL,
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	if u.URLService.DisableDB == "0" {
		existRedirect, _ := u.URLService.GetRedirectByURL(redirect.URL)
		if len(existRedirect.URL) > 0 {
			render.Status(r, http.StatusConflict)
			w.WriteHeader(http.StatusConflict)

			res := models.URLDataResponse{Result: u.URLService.MakeFullURL(existRedirect.Redirect)}
			render.JSON(w, r, res)
			return
		}
		_, err := u.URLService.NewRedirect(*redirect)
		if err != nil {
			return
		}

	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	res := models.URLDataResponse{Result: u.URLService.MakeFullURL(newRedirect)}
	render.JSON(w, r, res)
}

func (u *URLController) CreateBatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := &models.URLBatchData{}
	links, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	err = json.Unmarshal(links, &data)
	if err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	var redirects []*models.Redirect
	var jsonResults []models.URLBatchResponse

	for _, d := range *data {
		var resData models.URLBatchResponse
		newRedirect := helpers.GenerateRandomURL(10)
		redirect := &models.Redirect{
			ID:         d.CorrelationID,
			IsActive:   1,
			URL:        d.OriginalURL,
			Redirect:   newRedirect,
			DateCreate: time.Now().String(),
			DateUpdate: time.Now().String(),
		}
		resData.CorrelationID = d.CorrelationID
		resData.ShortURL = u.URLService.MakeFullURL(newRedirect)

		redirects = append(redirects, redirect)
		jsonResults = append(jsonResults, resData)
	}

	if u.URLService.DisableDB == "0" {
		_, err := u.URLService.NewRedirectsBatch(redirects)
		if err != nil {
			return
		}
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, jsonResults)
}

func (u *URLController) GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	id := chi.URLParam(r, "id")
	u.URLService.Logger.Warn().Str("id", id).Msg("GetHandler request")

	if len(id) > 0 {
		redirect, err := u.URLService.GetRedirect(id)
		if err != nil {
			u.URLService.Logger.Error().Err(err).Str("data", id).Msg("GetRedirect error")

			http.Error(w, "GetRedirect error", http.StatusBadRequest)
		}
		if redirect.Redirect != "" {
			fullRedirect := u.URLService.MakeFullURL(redirect.URL)
			u.URLService.Logger.Warn().Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect success")

			w.Header().Set("Location", fullRedirect)
			w.WriteHeader(http.StatusTemporaryRedirect)
			http.Redirect(w, r, fullRedirect, http.StatusTemporaryRedirect)
		} else {
			u.URLService.Logger.Error().Err(err).Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect not found")
		}
	} else {
		u.URLService.Logger.Error().Str("data", id).Msg("GetRedirect not found id empty")
		http.Error(w, "GetRedirect error", http.StatusBadRequest)
	}
}
