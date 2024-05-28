// Package websockets contains various websocket connections
package websockets

import "net/http"

// Websocket contains the interface that a Websocket should implment
type Websocket interface {
	Method() string
	Path() string
	Handler(http.ResponseWriter, *http.Request)
}
