// Package stream contains websockets that are related to the stream
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// WebSocket contains all the websockets that are related to the stream
func WebSocket(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()
	r.Get("/view/{topic}", func(w http.ResponseWriter, r *http.Request) {
		view(w, r, e, c)
	})

	r.Get("/add/{topic}", func(w http.ResponseWriter, r *http.Request) {
		add(w, r, e, c)
	})

	return r
}
