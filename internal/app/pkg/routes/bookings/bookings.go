// Package bookings contains routes that are related to bookings
package bookings

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

var (
	h = lib.WrapHandler
	m = lib.WrapMiddleware
)

// Router contains the bookings router
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/", func(r chi.Router) {
		r.Use(m(middlewares.IsAdminOrIsSuperAdmin, e, c))
		r.Get("/", h(index, e, c))
	})

	r.Get("/view/{booking_id}", h(view, e, c))

	return r
}
