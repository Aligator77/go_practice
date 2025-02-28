package controllers

import (
	"context"
	"net/http"

	"github.com/go-chi/render"
	"github.com/golang-migrate/migrate/v4"
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

func (d *DBController) Migrate(dsn *string) error {
	if d.DB.DisableDBStore == "0" && len(*dsn) > 0 {
		m, err := migrate.New("file:///migrations", *dsn)
		if err != nil {
			return err
		}
		err = m.Up()
		if err != nil {
			return err
		}
	}
	return nil
}
