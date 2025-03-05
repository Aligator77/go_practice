package controllers

import (
	"encoding/json"
	"errors"
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
	URLStore *stores.URLStore
}

func NewURLController(URLService *stores.URLStore) *URLController {
	return &URLController{
		URLStore: URLService,
	}
}

func (u *URLController) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	u.GetUserID(w, r)

	data, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	validateURL, err := helpers.ValidateURL(string(data))
	if !validateURL || err != nil {
		u.URLStore.Logger.Err(err).Msg("Write error CreatePostHandler")
		return
	}
	newRedirect := helpers.GenerateRandomURL(10)
	newUUID, _ := uuid.NewV7()

	redirect := &models.Redirect{
		ID:         newUUID.String(),
		IsDelete:   0,
		URL:        string(data),
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	existRedirect, _ := u.URLStore.GetRedirectByURL(redirect.URL)
	if len(existRedirect.URL) > 0 {
		render.Status(r, http.StatusConflict)
		w.WriteHeader(http.StatusConflict)

		_, err = w.Write([]byte(u.URLStore.MakeFullURL(existRedirect.Redirect)))
		if err != nil {
			u.URLStore.Logger.Err(err).Msg("Write error CreatePostHandler")
		}

		return
	}
	_, err = u.URLStore.NewRedirect(*redirect)
	if err != nil {
		return
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(u.URLStore.MakeFullURL(newRedirect)))
	if err != nil {
		return
	}
}

func (u *URLController) CreateRestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u.GetUserID(w, r)

	data := &models.URLData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}

	validateURL, err := helpers.ValidateURL(data.URL)
	if !validateURL || err != nil {
		u.URLStore.Logger.Err(err).Msg("Write error CreatePostHandler")
		return
	}
	newUUID, _ := uuid.NewV7()
	newRedirect := helpers.GenerateRandomURL(10)
	redirect := &models.Redirect{
		ID:         newUUID.String(),
		IsDelete:   0,
		URL:        data.URL,
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	existRedirect, _ := u.URLStore.GetRedirectByURL(redirect.URL)
	if len(existRedirect.URL) > 0 {
		render.Status(r, http.StatusConflict)
		w.WriteHeader(http.StatusConflict)

		res := models.URLDataResponse{Result: u.URLStore.MakeFullURL(existRedirect.Redirect)}
		render.JSON(w, r, res)
		return
	}
	_, err = u.URLStore.NewRedirect(*redirect)
	if err != nil {
		return
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	res := models.URLDataResponse{Result: u.URLStore.MakeFullURL(newRedirect)}
	render.JSON(w, r, res)
}

func (u *URLController) CreateBatchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	u.GetUserID(w, r)

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

		validateURL, err := helpers.ValidateURL(d.OriginalURL)
		if !validateURL || err != nil {
			u.URLStore.Logger.Err(err).Msg("Write error CreatePostHandler")
			return
		}
		redirect := &models.Redirect{
			ID:         d.CorrelationID,
			IsDelete:   0,
			URL:        d.OriginalURL,
			Redirect:   newRedirect,
			DateCreate: time.Now().String(),
			DateUpdate: time.Now().String(),
		}
		resData.CorrelationID = d.CorrelationID
		resData.ShortURL = u.URLStore.MakeFullURL(newRedirect)

		redirects = append(redirects, redirect)
		jsonResults = append(jsonResults, resData)
	}

	if u.URLStore.DisableDB == "0" {
		_, err := u.URLStore.NewRedirectsBatch(redirects)
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
	u.GetUserID(w, r)

	id := chi.URLParam(r, "id")
	u.URLStore.Logger.Warn().Str("id", id).Msg("GetHandler request")

	if len(id) > 0 {
		redirect, err := u.URLStore.GetRedirect(id)
		if err != nil {
			u.URLStore.Logger.Error().Err(err).Str("data", id).Msg("GetRedirect error")

			http.Error(w, "GetRedirect error", http.StatusBadRequest)
		}
		if redirect.Redirect != "" && redirect.IsDelete == 0 {
			fullRedirect := u.URLStore.MakeFullURL(redirect.URL)
			u.URLStore.Logger.Warn().Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect success")

			w.Header().Set("Location", fullRedirect)
			w.WriteHeader(http.StatusTemporaryRedirect)
			http.Redirect(w, r, fullRedirect, http.StatusTemporaryRedirect)
		} else if redirect.IsDelete == 1 {
			render.Status(r, http.StatusGone)
			w.WriteHeader(http.StatusGone)
			u.URLStore.Logger.Error().Err(err).Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect is deleted")
		} else {
			u.URLStore.Logger.Error().Err(err).Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect not found")
		}
	} else {
		u.URLStore.Logger.Error().Str("data", id).Msg("GetRedirect not found id empty")
		http.Error(w, "GetRedirect error", http.StatusBadRequest)
	}
}

func (u *URLController) CreateFullRestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	method := r.Method
	cookie, err := r.Cookie("user")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	u.GetUserID(w, r)

	switch method {
	case http.MethodGet:
		if len(cookie.String()) > 0 && len(cookie.Value) > 0 {
			existRedirects, err := u.URLStore.GetRedirectsByUser(cookie.Value)
			if err != nil {
				u.URLStore.Logger.Err(err).Str("cookie data", cookie.String()).Msg("error with cookie user")
			}
			if len(existRedirects) == 0 {
				render.Status(r, http.StatusNoContent)
				w.WriteHeader(http.StatusNoContent)
				return
			}
			var jsonResults []models.URLBatchResponse
			for _, e := range existRedirects {
				redirect := models.URLBatchResponse{
					ShortURL:    e.Redirect,
					OriginalURL: e.URL,
				}
				jsonResults = append(jsonResults, redirect)
			}
			render.JSON(w, r, jsonResults)
		}
		auth := r.Header.Get("Authorization")
		if len(cookie.String()) == 0 && len(auth) == 0 {
			render.Status(r, http.StatusUnauthorized)
			w.WriteHeader(http.StatusUnauthorized)
		}
	case http.MethodDelete:
		data, err := io.ReadAll(r.Body)
		if err != nil {
			_ = render.Render(w, r, server.ErrInvalidRequest(err))
			return
		}
		var urls []string
		json.Unmarshal(data, &urls)

		status, _ := u.URLStore.DeleteRedirect(urls)
		if status {
			render.Status(r, http.StatusAccepted)
			w.WriteHeader(http.StatusAccepted)
		}
	case http.MethodPost:
		data := &models.URLData{}
		if err := render.Bind(r, data); err != nil {
			_ = render.Render(w, r, server.ErrInvalidRequest(err))
			return
		}
		newUUID, _ := uuid.NewV7()
		newRedirectURL := helpers.GenerateRandomURL(10)
		newRedirect := &models.Redirect{
			ID:         newUUID.String(),
			IsDelete:   0,
			URL:        data.URL,
			Redirect:   newRedirectURL,
			DateCreate: time.Now().String(),
			DateUpdate: time.Now().String(),
		}
		existRedirect, _ := u.URLStore.GetRedirectByURL(newRedirect.URL)
		if len(existRedirect.URL) > 0 {
			render.Status(r, http.StatusConflict)
			w.WriteHeader(http.StatusConflict)

			res := models.URLDataResponse{Result: u.URLStore.MakeFullURL(existRedirect.Redirect)}
			render.JSON(w, r, res)
			return
		}
		_, err := u.URLStore.NewRedirect(*newRedirect)
		if err != nil {
			return
		}

	}

}

func (u *URLController) GetUserID(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("user")
	if err != nil && !errors.Is(err, http.ErrNoCookie) {
		u.URLStore.Logger.Err(err).Msg("error with cookie user")
		return
	}
	if len(cookie.String()) == 0 {
		expiration := time.Now().Add(365 * 24 * time.Hour)
		newUserID, _ := uuid.NewV7()
		newCookie := http.Cookie{Name: "user", Value: newUserID.String(), Expires: expiration}
		http.SetCookie(w, &newCookie)
		w.Header().Set("Authorization", newUserID.String())
	}
}
