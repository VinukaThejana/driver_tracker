package routes

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"go.uber.org/zap"
)

// Health is route that provided the health of the HTTP server
type Health struct {
	E *env.Env
	C *connections.C
	L *zap.SugaredLogger
}

// Method contains the method of the health route
func (health *Health) Method() string {
	return http.MethodGet
}

// Path contains the path of the health route
func (health *Health) Path() string {
	return "/health"
}

// Handler contains the bussiness logic of the health route
func (health *Health) Handler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	html := `
<!DOCTYPE html>
<html>
    <head>
        <title>Health Check</title>
    </head>
    <body>
        <h1>Everything is working as expected</h1>
    </body>
</html>
    `

	w.Write([]byte(html))
}
