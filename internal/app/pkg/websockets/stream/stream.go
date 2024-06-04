// Package stream contains websockets that are related to the stream
package stream

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Stream contains all the websockets that are related to the stream
func Stream(r chi.Router, e *env.Env, c *connections.C) {
	r.Route("/stream", func(r chi.Router) {
		r.Get("/view/{topic}", func(w http.ResponseWriter, r *http.Request) {
			view(w, r, e, c)
		})

		r.Get("/add/{topic}", func(w http.ResponseWriter, r *http.Request) {
			add(w, r, e, c)
		})
	})
}
