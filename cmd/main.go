// Spoton Cars streaming platfrom
package main

import (
	"fmt"
	"net/http"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/controllers"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/segmentio/kafka-go"
)

var (
	e         env.Env
	connector connections.C

	streamC controllers.Stream
)

func init() {
	e.Load()
	connector.InitRedis(&e)

	streamC = controllers.Stream{
		E: &e,
		C: &connector,
	}
}

func main() {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders: []string{"Content-Type", "X-CSRF-Token"},
	}))

	router.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "Keep-alive")

		streamC.Subscribe(w, kafka.LastOffset)
	})

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
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
	})

	logger.Log(fmt.Sprintf("Listening and running on port -> %d", e.Port))
	lib.LogFatal(http.ListenAndServe(fmt.Sprintf(":%d", e.Port), router))
}
