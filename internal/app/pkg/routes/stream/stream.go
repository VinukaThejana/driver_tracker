// Package stream contains all the routes that are related to streams
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router a route group that contains all the routes that are related to stream
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()
	r.Get("/view", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})
	r.Route("/create", func(r chi.Router) {
		r.Use(middlewares.IsContentJSON)
		r.Use(middlewares.IsDriver)
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			create(w, r, e, c)
		})
	})

	r.Route("/", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsBookingTokenValid(h, c, e)
		})
		r.Route("/add", func(r chi.Router) {
			r.Use(middlewares.IsContentJSON)
			r.Post("/{topic}", func(w http.ResponseWriter, r *http.Request) {
				add(w, r, e, c)
			})
		})
		r.Delete("/end", func(w http.ResponseWriter, r *http.Request) {
			end(w, r, e, c)
		})
	})

	return r
}
