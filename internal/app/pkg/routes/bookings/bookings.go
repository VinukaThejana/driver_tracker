// Package bookings contains routes that are related to bookings
package bookings

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains the bookings router
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Get("/view/{booking_id}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})

	return r
}
