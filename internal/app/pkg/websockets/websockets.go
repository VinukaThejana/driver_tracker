// Package websockets contains various websocket connections
package websockets

import (
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/websockets/stream"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Websocket contains all the websocket connections
type Websocket struct {
	E *env.Env
	C *connections.C
}

// Websocket contains all the websocket connections
func (w *Websocket) Websocket(r *chi.Mux) {
	r.Route("/ws", func(r chi.Router) {
		stream.Stream(r, w.E, w.C)
	})
}
