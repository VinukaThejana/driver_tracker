// Spoton Cars streaming platfrom
package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/controllers"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/services"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	e         env.Env
	connector connections.C
	log       *zap.SugaredLogger

	streamC controllers.Stream
)

func init() {
	e.Load()
	connector.InitRedis(&e)

	logger, err := zap.NewProduction()
	lib.LogFatal(err)
	defer logger.Sync()
	log = logger.Sugar()

	streamC = controllers.Stream{
		E: &e,
		C: &connector,
	}
}

func main() {
	errCh := make(chan error, 1)

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

	router.Get("/view", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "Keep-alive")

		streamC.Subscribe(w, e.KafkaTopic, kafka.LastOffset)
	})

	router.Get("/view/{topic}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "Keep-alive")

		offset := kafka.LastOffset

		topic := chi.URLParam(r, "topic")
		switch topic {
		case "":
			topic = e.KafkaTopic
		case "log":
			offset = kafka.LastOffset
		case "logs":
			topic = "log"
			offset = kafka.FirstOffset
		default:
		}

		streamC.Subscribe(w, topic, offset)
	})

	router.Post("/add/{topic}", func(w http.ResponseWriter, r *http.Request) {
		const maxRequestBodySize = 1 << 20
		w.Header().Set("Content-Type", "application/json")
		type Response map[string]interface{}

		contentType := r.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			w.WriteHeader(http.StatusUnsupportedMediaType)
			json.NewEncoder(w).Encode(Response{
				"status":  "bad_request",
				"message": "only content of type application/json can be sent",
			})
			go services.Log(&connector, &e, fmt.Sprintf("Sending invalid content type : %v", contentType))
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
		defer r.Body.Close()

		topic := chi.URLParam(r, "topic")
		if topic == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				"status":  "bad_request",
				"message": "topic is not provided",
			})
			go services.Log(&connector, &e, "topic is not provided")
			return
		}

		var data map[string]interface{}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			if errors.Is(err, io.EOF) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(Response{
					"status":  "bad_request",
					"message": "body cannot be empty, please provide valid json",
				})
				go services.Log(&connector, &e, "the provided request body is empty")
				return
			}

			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{
				"status":  "bad_request",
				"message": "failed to parse data invalid json",
			})
			go services.Log(&connector, &e, "invalid json object provided by the client")
			return
		}

		dataBytes, err := json.Marshal(data)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(Response{
				"status":  "internal_server_error",
				"message": "something went wrong on the server side",
			})
			go services.Log(&connector, &e, fmt.Sprintf("Error when marshaling the requst body\n%v", err))
			return
		}
		dataStr := string(dataBytes)

		connector.KafkaWriteToTopic(&e, topic, dataStr)
		go services.Log(&connector, &e, fmt.Sprintf("Logged %v", dataStr))

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{
			"status":  "okay",
			"message": "added the message to the topic",
		})
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

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	log.Infow("Started the HTTP server", zap.Int("address", e.Port))
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%d", e.Port), router); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("HTTP server error", zap.Error(err))
			errCh <- err
		}
	}()

	select {
	case sig := <-signalCh:
		log.Info("shutting down server", zap.String("signal", sig.String()))
	case err := <-errCh:
		log.Error("shutting down server", zap.Error(err))
	}
}
