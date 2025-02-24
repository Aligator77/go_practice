package stores

import (
	"encoding/json"
	"github.com/gofrs/uuid"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/models"
	"github.com/Aligator77/go_practice/internal/server"
)

type UrlService struct {
	Db         *helpers.ConnectionPool
	BaseUrl    string
	logger     log.Logger
	LocalStore string
}

func CreateUrlService(db *helpers.ConnectionPool, logger log.Logger, str string, localStore string) (us *UrlService) {
	us = &UrlService{
		Db:         db,
		BaseUrl:    str,
		logger:     logger,
		LocalStore: localStore,
	}

	return us
}

func (u *UrlService) Shutdown() error {
	err := u.Db.Close()

	return err
}

func (u *UrlService) MakeFullUrl(link string) string {
	if !strings.Contains(link, "http") && len(u.BaseUrl) > 0 {
		fullRedirect, _ := url.Parse(u.BaseUrl)
		fullRedirect.Path = link
		return fullRedirect.String()
	} else {
		return link
	}
}

func (u *UrlService) StoreToFile(link string) error {
	if len(u.LocalStore) > 0 {
		f, err := os.OpenFile(u.LocalStore, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
		if err != nil {
			level.Error(u.logger).Log("msg", "Cannot open localStoreFile", "err", err)
			return err
		}

		defer f.Close()

		if _, err = f.WriteString(link); err != nil {
			level.Error(u.logger).Log("msg", "Cannot write to localStoreFile", "err", err)
			return err
		}
	}
	return nil
}

func (u *UrlService) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	data, err := io.ReadAll(r.Body)
	if err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}
	newRedirect := helpers.GenerateRandomUrl(10)
	redirect := &models.Redirect{
		IsActive:   1,
		Url:        string(data),
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

	_, err = w.Write([]byte(u.MakeFullUrl(newRedirect)))
	if err != nil {
		return
	}
	//_ = render.Render(w, r, NewArticleResponse(article))
}

func (u *UrlService) CreateRestHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	data := &models.UrlData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}

	newRedirect := helpers.GenerateRandomUrl(10)
	redirect := &models.Redirect{
		IsActive:   1,
		Url:        data.Url,
		Redirect:   newRedirect,
		DateCreate: time.Now().String(),
		DateUpdate: time.Now().String(),
	}
	_, err := u.NewRedirect(*redirect)
	if err != nil {
		return
	}

	//{"uuid":"1","short_url":"4rSPg8ap","original_url":"http://yandex.ru"}
	newUuid, _ := uuid.NewV7()
	redirect.ID = newUuid.String()
	dataFile, _ := json.Marshal(redirect)
	_ = u.StoreToFile(string(dataFile))

	render.JSON(w, r, u.MakeFullUrl(newRedirect))

	//_ = render.Render(w, r, NewArticleResponse(article))
}

func (u *UrlService) GetRedirect(id string) (redirect models.Redirect, err error) {

	sqlRequest, ctx, cancel := Get(GetRedirect)
	defer cancel()

	conn, err := u.Db.Conn(ctx)
	if err != nil {
		level.Error(u.logger).Log("msg", "GetRedirect get connection failure", "err", err)
		return redirect, err
	}
	defer conn.Close()

	row, err := conn.QueryContext(ctx, sqlRequest, id)
	if err != nil {
		level.Error(u.logger).Log(
			"msg", "GetRedirect exec failure",
			"err", err,
			"data", id,
		)

		return redirect, err
	}

	for row.Next() {

		if err := row.Scan(
			&redirect.Url,
			&redirect.Redirect,
			&redirect.DateCreate,
			&redirect.DateUpdate,
		); err != nil {
			level.Warn(u.logger).Log("scan failure", "err", err)
			continue
		}
	}

	return redirect, nil
}

func (u *UrlService) NewRedirect(redirect models.Redirect) (id int64, err error) {
	sqlRequest, ctx, cancel := Get(InsertRedirects)
	defer cancel()

	conn, err := u.Db.Conn(ctx)
	if err != nil {
		level.Error(u.logger).Log("msg", "NewRedirect get connection failure", "err", err)
		return id, err
	}
	defer conn.Close()

	res, err := conn.ExecContext(ctx, sqlRequest, redirect.IsActive, redirect.Url, redirect.Redirect)
	if err != nil {
		level.Error(u.logger).Log(
			"msg", "NewRedirect exec failure",
			"err", err,
			"data", id,
		)

		return id, err
	}

	if affected, err := res.RowsAffected(); affected > 0 {
		level.Warn(u.logger).Log("msg", "NewRedirect exec has affected rows", "affected", affected)
	} else if err != nil {
		level.Error(u.logger).Log(
			"msg", "NewRedirect exec failure",
			"err", err,
			"data", id,
		)
	}

	return id, nil
}

func (u *UrlService) GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	id := chi.URLParam(r, "id")
	if len(id) > 0 {
		redirect, err := u.GetRedirect(id)
		if err != nil {
			level.Error(u.logger).Log(
				"msg", "GetRedirect error",
				"err", err,
				"data", id,
			)
			http.Error(w, "GetRedirect error", http.StatusBadRequest)
		}
		if redirect.Redirect != "" {
			fullRedirect := u.MakeFullUrl(redirect.Url)

			level.Warn(u.logger).Log("msg", "GetRedirect success", "redirect.Url", redirect.Url, "redirect.Redirect", redirect.Redirect)
			w.Header().Set("Location", fullRedirect)
			w.WriteHeader(http.StatusTemporaryRedirect)
			http.Redirect(w, r, fullRedirect, http.StatusTemporaryRedirect)
		} else {
			level.Error(u.logger).Log("msg", "GetRedirect not found", "id", id, "redirect.Url", redirect.Url, "redirect.Redirect", redirect.Redirect)
		}
	} else {
		level.Error(u.logger).Log("msg", "GetRedirect not found id empty", "id", id)
		http.Error(w, "GetRedirect error", http.StatusBadRequest)
	}
}
