// Package index contains routes that does not belong to a collection
package index

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Index contains all the routes that do not have a collection
func Index(r chi.Router, e *env.Env) {
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, e.ApiDoc, http.StatusMovedPermanently)
	})
	r.Get("/health", health)
}
