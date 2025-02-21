package stores

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/Aligator77/go_practice/internal/helpers"
	"github.com/Aligator77/go_practice/internal/models"
	"github.com/Aligator77/go_practice/internal/server"
)

type UrlService struct {
	Db       *helpers.ConnectionPool
	SiteHost string
	logger   log.Logger
}

func CreateUrlService(db *helpers.ConnectionPool, logger log.Logger, str string) (us *UrlService) {
	us = &UrlService{
		Db:       db,
		SiteHost: str,
		logger:   logger,
	}

	return us
}

func (u *UrlService) Shutdown() error {
	err := u.Db.Close()

	return err
}

func (u *UrlService) GetRedirectsResponse() error {
	err := u.Db.Close()

	return err
}

func (u *UrlService) CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	data := &models.UrlData{}
	if err := render.Bind(r, data); err != nil {
		_ = render.Render(w, r, server.ErrInvalidRequest(err))
		return
	}

	//url := data.Url
	//u.Db.NewRedirect(url)

	render.Status(r, http.StatusCreated)
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

	row, err := conn.ExecContext(ctx, sqlRequest, sql.Named("Data", string(id)))
	if err != nil {
		level.Error(u.logger).Log(
			"msg", "GetRedirect exec failure",
			"err", err,
			"data", string(id),
		)

		return redirect, err
	}

	for row.Next() {

		if err := row.Scan(
			&redirect.Redirect,
			&redirect.Url,
			&redirect.DateCreate,
			&redirect.DateUpdate,
		); err != nil {
			level.Warn(u.logger).Log("scan failure", "err", err)
			continue
		}
	}

	return redirect, nil
}
