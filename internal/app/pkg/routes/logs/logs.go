// Package log contains the routes related to log dumps related to various jobs
package logs

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains all the routes that do not have a collection
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})
	r.Delete("/delete/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		delete(w, r, e, c)
	})

	return r
}
