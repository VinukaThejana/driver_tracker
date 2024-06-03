// Package routes contains all the routes
package routes

import (
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

// Route interface consits methods that should be implmented by a specific route
type Route interface {
	Method() string
	Path() string
	Handler(http.ResponseWriter, *http.Request)
}

type response map[string]interface{}

type RouteType uint8

const (
	// RouteTypeAdd is the add route
	RouteTypeAdd RouteType = iota
	// RouteTypeCreate is the create route
	RouteTypeCreate
	// RouteTypeEnd is the end route
	RouteTypeEnd
	// RouteTypeHealth is the health route
	RouteTypeHealth
)

// NewConfig is a config contaning all the routes
func NewConfig(e *env.Env, c *connections.C) map[RouteType]Route {
	return map[RouteType]Route{
		RouteTypeAdd: &Add{
			E: e,
			C: c,
		},
		RouteTypeCreate: &CreateStream{
			E: e,
			C: c,
		},
		RouteTypeEnd: &EndStream{
			E: e,
			C: c,
		},
		RouteTypeHealth: &Health{
			E: e,
			C: c,
		},
	}
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	sonic.ConfigDefault.NewEncoder(w).Encode(response{
		"message": message,
	})
}

func sendJSONResponseWInterface(w http.ResponseWriter, statusCode int, res map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	sonic.ConfigDefault.NewEncoder(w).Encode(res)
}
