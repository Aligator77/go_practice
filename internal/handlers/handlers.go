package handlers

import (
	"github.com/Aligator77/go_practice/internal/stores"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	// пока установим ответ-заглушку, без проверки ошибок
	_, _ = w.Write([]byte(`
      {
        "response": {
          "text": "OK"
        },
        "version": "1.0"
      }
    `))
}

func GetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")

	id := chi.URLParam(r, "id")
	if len(id) > 0 {
		redirect := stores.GetRedirectsResponse(id)

		http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	} else {

	}
}
