package routes

import (
	"net/http"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/controllers"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/segmentio/kafka-go"
)

// View is used to view messages on a given topic
type View struct {
	E *env.Env
	C *connections.C
}

// Method is a used to get the Method of the view route
func (view *View) Method() string {
	return http.MethodGet
}

// Path is used to get the Path of the view route
func (view *View) Path() string {
	return "/view/{topic}"
}

// Handler is used to get the Handler of the view route
func (view *View) Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "Keep-alive")

	offset := kafka.LastOffset

	topic := chi.URLParam(r, "topic")
	switch topic {
	case "":
		topic = view.E.KafkaTopic
	case "log":
		offset = kafka.LastOffset
	case "logs":
		topic = "log"
		offset = kafka.FirstOffset
	default:
	}

	streamC := controllers.Stream{
		E: view.E,
		C: view.C,
	}

	streamC.Subscribe(w, topic, offset)
}
