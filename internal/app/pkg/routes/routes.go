// Package routes contains all the routes
package routes

import (
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/routes/index"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/routes/stream"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
)

// Route contains various third party connections and eviromental variables that are used throughout the routes
type Route struct {
	E *env.Env
	C *connections.C
}

// Routes contains all the HTTP routes
func (route *Route) Routes(r *chi.Mux) {
	index.Index(r, route.E)
	stream.Stream(r, route.E, route.C)
}
