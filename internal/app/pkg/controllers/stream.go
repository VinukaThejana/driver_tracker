package controllers

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
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

	mechanism, _ := scram.Mechanism(scram.SHA512, s.E.KafkaUsername, s.E.KafkaPassword)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{s.E.KafkaBroker},
		GroupID:     uuid.New().String(),
		Topic:       topic,
		StartOffset: offset,
		Dialer: &kafka.Dialer{
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		},
	})
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
