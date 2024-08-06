// Package jobs is a package that contains routes that are related to jobs
package jobs

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

// Router is a function that contains routes that are related to jobs
func Router(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Route("/end", func(r chi.Router) {
		r.Use(m(middlewares.IsAdminOrIsSuperAdmin, e, c))
		r.Delete("/{booking_id}", h(end, e, c))
	})

	r.Route("/reset", func(r chi.Router) {
		r.Use(m(middlewares.IsSuperAdmin, e, c))
		r.Delete("/", h(reset, e, c))
	})

	r.Route("/", func(r chi.Router) {
		r.Use(m(middlewares.IsCron, e, c))
		r.Patch("/rotate", h(rotate, e, c))
		r.Patch("/check_job", h(checkJob, e, c))
	})

	return r
}
