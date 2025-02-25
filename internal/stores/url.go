package stores

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/models"
	"github.com/Aligator77/go_practice/internal/server"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type URLService struct {
	DB         *helpers.ConnectionPool
	BaseURL    string
	logger     zerolog.Logger
	LocalStore string
	DisableDB  string
	EmulateDB  map[string]models.Redirect
}

func CreateURLService(db *helpers.ConnectionPool, logger zerolog.Logger, BaseURL string, localStore string, DisableDBStore string) (us *URLService) {
	us = &URLService{
		DB:         db,
		BaseURL:    BaseURL,
		logger:     logger,
		LocalStore: localStore,
		DisableDB:  DisableDBStore,
		EmulateDB:  make(map[string]models.Redirect, 0),
	}

	return us
}

func (u *URLService) Shutdown() error {
	err := u.DB.Close()

	return err
}

func (u *URLService) MakeFullURL(link string) string {
	if !strings.Contains(link, "http") && len(u.BaseURL) > 0 {
		fullRedirect, _ := url.Parse(u.BaseURL)
		fullRedirect.Path = link
		return fullRedirect.String()
	} else {
		return link
	}
}

func (u *URLService) StoreToFile(link string) error {
	if len(u.LocalStore) > 0 {
		f, err := os.OpenFile(u.LocalStore, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			u.logger.Error().Err(err).Msg("Cannot open localStoreFile")
			return err
		}

		defer f.Close()

		if _, err = f.WriteString(link); err != nil {
			u.logger.Error().Err(err).Msg("Cannot write to localStoreFile")
			return err
		}
	}
	return nil
}

func (u *URLService) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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
	_, err = u.NewRedirect(*redirect)
	if err != nil {
		return
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)

	_, err = w.Write([]byte(u.MakeFullURL(newRedirect)))
	if err != nil {
		return
	}
	//_ = render.Render(w, r, NewArticleResponse(article))
}

func (u *URLService) CreateRestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := &models.URLData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}

	newRedirect := helpers.GenerateRandomURL(10)
	redirect := &models.Redirect{
		IsActive:   1,
		URL:        data.URL,
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	if u.DisableDB == "0" {

		_, err := u.NewRedirect(*redirect)
		if err != nil {
			return
		}
	}
	//{"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}
	newUUID, _ := uuid.NewV7()
	redirect.ID = newUUID.String()
	dataFile, _ := json.Marshal(redirect)
	_ = u.StoreToFile(string(dataFile))

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, u.MakeFullURL(newRedirect))
	u.EmulateDB[redirect.Redirect] = *redirect
	//_ = render.Render(w, r, NewArticleResponse(article))
}

func (u *URLService) GetRedirect(id string) (redirect models.Redirect, err error) {
	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(GetRedirect)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.logger.Error().Err(err).Msg("GetRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		row, err := conn.QueryContext(ctx, sqlRequest, id)
		if err != nil {
			u.logger.Error().Err(err).Str("data", id).Msg("GetRedirect exec failure")
			return redirect, err
		}

		for row.Next() {

			if err := row.Scan(
				&redirect.URL,
				&redirect.Redirect,
				&redirect.DateCreate,
				&redirect.DateUpdate,
			); err != nil {
				u.logger.Error().Err(err).Msg("scan failure")
				continue
			}
		}

	} else {
		redirect = u.EmulateDB[id]
	}

	return redirect, nil
}

func (u *URLService) NewRedirect(redirect models.Redirect) (id int64, err error) {

	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(InsertRedirects)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.logger.Error().Err(err).Msg("NewRedirect get connection failure")
			return id, err
		}
		defer conn.Close()

		res, err := conn.ExecContext(ctx, sqlRequest, redirect.IsActive, redirect.URL, redirect.Redirect)
		if err != nil {
			u.logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirect get connection failure")
			return id, err
		}

		if affected, err := res.RowsAffected(); affected > 0 {
			u.logger.Warn().Str("affected", strconv.FormatInt(affected, 10)).Msg("NewRedirect exec has affected rows")
		} else if err != nil {
			u.logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirect get connection failure")
		}

	} else {
		u.EmulateDB[redirect.Redirect] = redirect
	}

	return id, nil
}

func (u *URLService) GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	id := chi.URLParam(r, "id")
	if len(id) > 0 {
		redirect, err := u.GetRedirect(id)
		if err != nil {
			u.logger.Error().Err(err).Str("data", id).Msg("GetRedirect error")

			http.Error(w, "GetRedirect error", http.StatusBadRequest)
		}
		if redirect.Redirect != "" {
			fullRedirect := u.MakeFullURL(redirect.URL)
			u.logger.Warn().Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect success")

			w.Header().Set("Location", fullRedirect)
			w.WriteHeader(http.StatusTemporaryRedirect)
			http.Redirect(w, r, fullRedirect, http.StatusTemporaryRedirect)
		} else {
			u.logger.Error().Err(err).Strs("data", []string{id, redirect.URL, redirect.Redirect}).Msg("GetRedirect not found")
		}
	} else {
		u.logger.Error().Str("data", id).Msg("GetRedirect not found id empty")
		http.Error(w, "GetRedirect error", http.StatusBadRequest)
	}
}
