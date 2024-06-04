package index

import (
	"net/http"
)

// health is a route that is used to check the health of the server
func health(w http.ResponseWriter, _ *http.Request) {
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
