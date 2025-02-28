package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/Aligator77/go_practice/internal/config"
)

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
		driver, _ := postgres.WithInstance(d.DB.DB(), &postgres.Config{})
		m, _ := migrate.NewWithDatabaseInstance(
			"file:///migrations",
			"postgres", driver)
		err := m.Up()
		if err != nil {
			return err
		}
	}
	return nil
}
