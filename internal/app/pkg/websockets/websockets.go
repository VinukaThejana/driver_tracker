// Package websockets contains various websocket connections
package websockets

import (
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/websockets/stream"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Websocket contains all the websocket connections
type Websocket struct {
	E *env.Env
	C *connections.C
}

// Websocket contains all the websocket connections
func (w *Websocket) Websocket() http.Handler {
	r := chi.NewRouter()
	r.Mount("/stream", stream.WebSocket(w.E, w.C))
	return r
}
