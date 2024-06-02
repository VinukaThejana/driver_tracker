package routes

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
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
	if topic == "" {
		http.Error(w, "subscribe to a valid booking id", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusBadRequest)
		return
	}

	reader := view.C.KafkaReader(view.E, topic, offset)
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()

	for {
		message, _ := reader.ReadMessage(ctx)
		payload, err := lib.ToStr(string(message.Value))
		if err != nil {
			log.Error().Err(err).Str("value", string(message.Value)).Msg("Error occured when serialization and deserialization")
			http.Error(w, "something went wrong", http.StatusInternalServerError)
			return
		}
		payloadStr := string(payload)

		data := fmt.Sprintf("data: %s\n\n", payloadStr)
		log.Info().Str("data", data)
		fmt.Fprint(w, data)

		flusher.Flush()
	}
}
