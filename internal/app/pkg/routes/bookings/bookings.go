// Package bookings contains routes that are related to bookings
package bookings

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains the bookings router
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsAdminOrIsSuperAdmin(h, e, c)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			index(w, r, e, c)
		})
	})

	r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})

	return r
}
