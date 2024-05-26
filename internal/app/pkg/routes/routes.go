// Package routes contains all the routes
package routes

import (
	"encoding/json"
	"net/http"
)

type response map[string]interface{}

// Route interface consits methods that should be implmented by a specific route
type Route interface {
	Method() string
	Path() string
	Handler(http.ResponseWriter, *http.Request)
}

func sendJSONResponse(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response{
		"message": message,
	})
}

func sendJSONResponseWInterface(w http.ResponseWriter, statusCode int, res interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(res)
}
