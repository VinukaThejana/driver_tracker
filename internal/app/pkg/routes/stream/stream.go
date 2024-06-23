// Package stream contains all the routes that are related to streams
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router a route group that contains all the routes that are related to stream
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()
	r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})
	r.Route("/create", func(r chi.Router) {
		r.Use(middlewares.IsContentJSON)
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsDriver(h, e, c)
		})
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			create(w, r, e, c)
		})
	})

	r.Route("/add", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsBookingTokenValid(h, e, c)
		})
		r.Use(middlewares.IsContentJSON)
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			add(w, r, e, c)
		})
	})
	r.Route("/end", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsBookingTokenValid(h, e, c)
		})
		r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
			end(w, r, e, c)
		})
	})

	return r
}
