// Package jobs is a package that contains routes that are related to jobs
package jobs

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router is a function that contains routes that are related to jobs
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/end", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsAdminOrIsSuperAdmin(h, e, c)
		})
		r.Delete("/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
			end(w, r, e, c)
		})
	})
	r.Route("/reset", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsSuperAdmin(h, e, c)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			reset(w, r, e, c)
		})
	})
	r.Patch("/rotate", func(w http.ResponseWriter, r *http.Request) {
		rotate(w, r, e, c)
	})

	return r
}
