package controllers

import (
	"context"
	"github.com/Aligator77/go_practice/internal/config"
	"github.com/go-chi/render"
	"net/http"
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

	return
}
