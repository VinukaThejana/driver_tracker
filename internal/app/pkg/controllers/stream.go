package controllers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

// Stream contains controllers related to streaming and subscribing to streaming services
type Stream struct {
	E *env.Env
	C *connections.C
}

// Subscribe is a function that is used to subscribe to the Kafka stream from a given Offset
func (s *Stream) Subscribe(w http.ResponseWriter, topic string, offset int64) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		log.Error().Msg("failed to cast the type to flusher")
		return
	}

	reader := s.C.KafkaReader(s.E, topic, offset)
	defer reader.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()

	for {
		message, _ := reader.ReadMessage(ctx)
		payload, err := lib.ToStr(string(message.Value))
		if err != nil {
			log.Error().Err(err).Str("value", string(message.Value)).Msg("Error occured when serialization and deserialization")
			return
		}
		payloadStr := string(payload)

		data := fmt.Sprintf("data: %s\n\n", payloadStr)
		log.Info().Str("data", data)
		fmt.Fprint(w, data)

		flusher.Flush()
	}
}
