// Package websockets contains various websocket connections
package websockets

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

// Websocket contains the interface that a Websocket should implment
type Websocket interface {
	Method() string
	Path() string
	Handler(http.ResponseWriter, *http.Request)
}

type WebsocketType uint8

const (
	// WebsocketTypeView is the view websocket
	WebsocketTypeView WebsocketType = iota
	// WebsocketTypeAdd is the add websocket
	WebsocketTypeAdd
)

// NewConfig creates a new config for Websocket connections
func NewConfig(e *env.Env, c *connections.C) map[WebsocketType]Websocket {
	return map[WebsocketType]Websocket{
		WebsocketTypeAdd: &Add{
			E: e,
			C: c,
		},
		WebsocketTypeView: &View{
			E: e,
			C: c,
		},
	}
}
