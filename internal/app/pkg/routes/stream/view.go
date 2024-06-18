package stream

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/errors"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// View is a route that is used to listen to the stream
func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "Keep-alive")

	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		http.Error(w, "please provide a valid booking id", http.StatusBadRequest)
		return
	}
	val := c.R.DB.Get(r.Context(), bookingID).Val()
	if val == "" {
		http.Error(w, "please provide a valid booking id", http.StatusBadRequest)
		return
	}
	partitionNo, err := strconv.Atoi(val)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert the partition number from string to int")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming is not supported", http.StatusBadRequest)
		return
	}

	reader := c.KafkaReader(e, e.Topic, partitionNo, kafka.LastOffset)
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()

	for {
		message, _ := reader.ReadMessage(ctx)
		payload, err := lib.ToStr(string(message.Value))
		if err != nil {
			log.Error().Err(err).Str("value", string(message.Value)).Msg("Error occured when serialization and deserialization")
			http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
			return
		}
		payloadStr := string(payload)

		data := fmt.Sprintf("data: %s\n\n", payloadStr)
		log.Info().Str("data", data)
		fmt.Fprint(w, data)

		flusher.Flush()
	}
}
