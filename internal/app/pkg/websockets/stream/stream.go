// Package stream contains websockets that are related to the stream
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

// WebSocket contains all the websockets that are related to the stream
func WebSocket(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()
	r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})

	r.Route("/add", func(r chi.Router) {
		r.Use(func(h http.Handler) http.Handler {
			return middlewares.IsBookingTokenValid(h, e, c)
		})
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			add(w, r, e, c)
		})
	})

	return r
}
