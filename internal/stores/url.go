package stores

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/gofrs/uuid"
	"github.com/rs/zerolog"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/models"
	"github.com/Aligator77/go_practice/internal/server"
)

type URLService struct {
	DB         *config.ConnectionPool
	BaseURL    string
	logger     zerolog.Logger
	LocalStore string
	DisableDB  string
	EmulateDB  map[string]models.Redirect
}

func CreateURLService(db *config.ConnectionPool, logger zerolog.Logger, BaseURL string, localStore string, DisableDBStore string) (us *URLService) {
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

		defer func(f *os.File) {
			err := f.Close()
			if err != nil {
				u.logger.Error().Err(err).Msg("Cannot close localStoreFile")
			}
		}(f)

		if _, err = f.WriteString(link); err != nil {
			u.logger.Error().Err(err).Msg("Cannot write to localStoreFile")
			return err
		}
	}
	return nil
}

func (u *URLService) CreatePostHandler(w http.ResponseWriter, r *http.Request) {

	if slices.Contains(r.Header.Values("Content-Encoding"), "gzip") {
		gzipReader, err := gzip.NewReader(r.Body)
		if err != nil {
			u.logger.Error().Err(err).Msg("failed to create gzip reader")
		}
		defer func(gzipReader *gzip.Reader) {
			err := gzipReader.Close()
			if err != nil {
				u.logger.Error().Err(err).Msg("failed to close gzip reader")
			}
		}(gzipReader)

		r.Body = io.NopCloser(gzipReader)
	}

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

	res := models.URLDataResponse{Result: u.MakeFullURL(newRedirect)}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, res)
	u.EmulateDB[redirect.Redirect] = *redirect
}

func (u *URLService) CreateBatchHandler(w http.ResponseWriter, r *http.Request) {
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
			Redirect:   u.MakeFullURL(newRedirect),
			DateCreate: time.Now().String(),
			DateUpdate: time.Now().String(),
		}
		resData.CorrelationID = d.CorrelationID
		resData.ShortURL = newRedirect

		dataFile, _ := json.Marshal(redirect)
		_ = u.StoreToFile(string(dataFile))
		u.EmulateDB[redirect.Redirect] = *redirect
		redirects = append(redirects, redirect)
		jsonResults = append(jsonResults, resData)
	}

	if u.DisableDB == "0" {
		_, err := u.NewRedirectsBatch(redirects)
		if err != nil {
			return
		}
	}

	render.Status(r, http.StatusCreated)
	w.WriteHeader(http.StatusCreated)
	render.JSON(w, r, jsonResults)
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

func (u *URLService) GetRedirectByURL(url string) (redirect models.Redirect, err error) {
	if u.DisableDB == "0" {
		sqlRequest, ctx, cancel := Get(GetRedirectByURL)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.logger.Error().Err(err).Msg("GetRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		row, err := conn.QueryContext(ctx, sqlRequest, url)
		if err != nil {
			u.logger.Error().Err(err).Str("data", url).Msg("GetRedirect exec failure")
			return redirect, err
		}

		for row.Next() {

			if err := row.Scan(
				&redirect.ID,
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
		redirect = u.EmulateDB[url]
	}

	return redirect, nil
}

func (u *URLService) NewRedirect(redirect models.Redirect) (res models.Redirect, err error) {
	newUUID, _ := uuid.NewV7()
	redirect.ID = newUUID.String()

	if u.DisableDB == "0" {
		existRedirect, _ := u.GetRedirectByURL(redirect.URL)
		if len(existRedirect.URL) > 0 {
			return redirect, err
		}
		sqlRequest, ctx, cancel := Get(InsertRedirect)
		defer cancel()

		conn, err := u.DB.Conn(ctx)
		if err != nil {
			u.logger.Error().Err(err).Msg("NewRedirect get connection failure")
			return redirect, err
		}
		defer conn.Close()

		res, err := conn.ExecContext(ctx, sqlRequest, redirect.ID, redirect.IsActive, redirect.URL, redirect.Redirect)
		if err != nil {
			u.logger.Error().Err(err).Str("data", redirect.String()).Msg("NewRedirect get connection failure")
			return redirect, err
		}

		if affected, err := res.RowsAffected(); affected > 0 {
			u.logger.Warn().Str("affected", strconv.FormatInt(affected, 10)).Msg("NewRedirect exec has affected rows")
		} else if err != nil {
			u.logger.Error().Err(err).Str("data", redirect.String()).Msg("NewRedirect get connection failure")
		}

	} else {
		if existRedirect, ok := u.EmulateDB[redirect.URL]; ok {
			return existRedirect, nil
		}
		u.EmulateDB[redirect.Redirect] = redirect
	}

	dataFile, _ := json.Marshal(redirect)
	_ = u.StoreToFile(string(dataFile))

	return redirect, nil
}

func (u *URLService) NewRedirectsBatch(redirects []*models.Redirect) (id int64, err error) {
	if u.DisableDB != "0" || len(redirects) == 0 {
		return 0, nil
	}

	sqlRequest, ctx, cancel := Get(InsertBatchRedirects)
	defer cancel()

	var queryStr strings.Builder
	queryStr.WriteString(sqlRequest)

	for i, r := range redirects {
		queryStr.WriteString(" (")
		queryStr.WriteString(`'` + r.ID + `', B'` + strconv.Itoa(r.IsActive) + `', '` + r.URL + `', '` + r.Redirect + `', NOW(), NOW()`)
		queryStr.WriteString(")")
		if i != len(redirects)-1 {
			queryStr.WriteString(",")
		}
	}

	conn, err := u.DB.Conn(ctx)
	if err != nil {
		u.logger.Error().Err(err).Msg("NewRedirect get connection failure")
		return id, err
	}
	defer conn.Close()
	res, err := conn.ExecContext(ctx, queryStr.String())
	if err != nil {
		u.logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirect get connection failure")
		return id, err
	}

	if affected, err := res.RowsAffected(); affected > 0 {
		u.logger.Warn().Str("affected", strconv.FormatInt(affected, 10)).Msg("NewRedirect exec has affected rows")
	} else if err != nil {
		u.logger.Error().Err(err).Str("data", strconv.FormatInt(id, 10)).Msg("NewRedirect get connection failure")
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
