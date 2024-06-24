// Package logs contains the routes related to log dumps related to various jobs
package logs

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains all the routes that do not have a collection
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsAdmin(h, e, c)
		})
		r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
			view(w, r, e, c)
		})
		r.Delete("/delete/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
			delete(w, r, e, c)
		})
	})

	return r
}
