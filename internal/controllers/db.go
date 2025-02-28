package controllers

import (
	"context"
	"embed"
	"github.com/go-chi/render"
	"github.com/pressly/goose/v3"
	"net/http"

	"github.com/Aligator77/go_practice/internal/config"
)

var embedMigrations embed.FS

type DBController struct {
	DB      *config.ConnectionPool
	Context context.Context
}

func NewDBController(db *config.ConnectionPool, ctx context.Context) *DBController {
	return &DBController{
		DB:      db,
		Context: ctx,
	}
}

func (d *DBController) CheckConnectHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	status := d.DB.CheckConnection(d.Context)
	if !status {
		render.Status(r, http.StatusInternalServerError)
		w.WriteHeader(http.StatusInternalServerError)
	}

}

func (d *DBController) Migrate() error {
	if d.DB.DisableDBStore == "0" {

		goose.SetBaseFS(embedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			return err
		}

		if err := goose.Up(d.DB.DB(), "migrations"); err != nil {
			return err
		}
	}
	return nil
}
