// Package index contains routes that does not belong to a collection
package index

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Router contains all the routes that do not have a collection
func Router(e *env.Env) http.Handler {
	r := chi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, e.ApiDoc, http.StatusMovedPermanently)
	})
	r.Get("/health", health)

	return r
}
