// Package routes contains all the routes
package routes

import "net/http"

// Route interface consits methods that should be implmented by a specific route
type Route interface {
	Method() string
	Path() string
	Handler(http.ResponseWriter, *http.Request)
}
