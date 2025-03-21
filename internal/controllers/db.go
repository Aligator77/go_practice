// Package controllers provide migrations and health handler
package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"github.com/pressly/goose/v3"

	"github.com/Aligator77/go_practice/internal/config"
	"github.com/Aligator77/go_practice/migrations"
)

type DBController struct {
	DB      *config.ConnectionPool
	Context context.Context
}

func NewDBController(ctx context.Context, db *config.ConnectionPool) *DBController {
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

func (d *DBController) Migrate(ctx context.Context) (result []*goose.MigrationResult, err error) {
	if d.DB.DisableDBStore == "0" {
		provider, err := goose.NewProvider("postgres", d.DB.DB(), migrations.Embed)
		if err != nil {
			return nil, err
		}

		result, err = provider.Up(ctx)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}
