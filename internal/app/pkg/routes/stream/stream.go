// Package stream contains all the routes that are related to streams
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

var (
	v = validator.New()
	h = lib.WrapHandler
	m = lib.WrapMiddleware
)

// Router a route group that contains all the routes that are related to stream
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/create", func(r chi.Router) {
		r.Use(middlewares.IsContentJSON)
		r.Use(m(middlewares.IsDriver, e, c))
		r.Post("/", h(create, e, c))
	})

	r.Route("/add", func(r chi.Router) {
		r.Use(middlewares.IsContentJSON)
		r.Use(m(func(h http.Handler, e *env.Env, c *connections.C) http.Handler {
			return middlewares.IsBookingTokenValid(h, e, c, false)
		}, e, c))
		r.Post("/", h(add, e, c))
		r.Post("/v2", h(addV2, e, c))
	})

	r.Route("/end", func(r chi.Router) {
		r.Use(m(middlewares.ValidateDriverOrBookingToken, e, c))
		r.Delete("/", h(end, e, c))
	})

	return r
}
