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
