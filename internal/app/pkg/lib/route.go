package lib

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
)

// WrapHandler is used to simplify route handlers
func WrapHandler(
	h func(http.ResponseWriter, *http.Request, *env.Env, *connections.C),
	e *env.Env,
	c *connections.C,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h(w, r, e, c)
	}
}

// WrapMiddleware is used to simplify middlewares
func WrapMiddleware(
	m func(http.Handler, *env.Env, *connections.C) http.Handler,
	e *env.Env,
	c *connections.C,
) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return m(h, e, c)
	}
}
