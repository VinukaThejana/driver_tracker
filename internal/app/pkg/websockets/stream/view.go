package stream

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	topic := chi.URLParam(r, "topic")
	if topic == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}

	isDriver := true

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	if isDriver {
		upgrader.OnMessage(func(_ *websocket.Conn, _ websocket.MessageType, b []byte) {
			var (
				data    map[string]interface{}
				payload string
				err     error
			)

			if err = sonic.UnmarshalString(string(b), &data); err != nil {
				log.Error().Err(err).Msg("provide valid JSON data")
				return
			}
			data["timestamp"] = time.Now().UTC().Unix()
			if payload, err = sonic.MarshalString(data); err != nil {
				log.Error().Err(err).Msg("failed to marshal data")
				return
			}

			if err = c.KafkaWriteToTopicWithHTTP(e, topic, payload); err != nil {
				log.Error().Err(err).Msg("failed to write data to kafka")
				return
			}
		})
	}

	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Info().Str("addr", conn.RemoteAddr().String()).Msg("connection opened")
		done := make(chan struct{})

		go func() {
			reader := c.KafkaReader(e, topic, kafka.LastOffset)
			defer reader.Close()
			defer close(done)

			for {
				select {
				case <-done:
					return
				default:
					message, _ := reader.ReadMessage(r.Context())
					if len(message.Value) == 0 {
						continue
					}
					if err := conn.WriteMessage(websocket.TextMessage, message.Value); err != nil {
						log.Error().Err(err).Msg("error sending data to the websocket client")
						return
					}
				}
			}
		}()

		conn.OnClose(func(c *websocket.Conn, err error) {
			if err != nil {
				log.Error().Err(err).Str("addr", c.RemoteAddr().String()).Msg("connection closed with error")
			} else {
				log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection closed")
			}
		})
	})

	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("error occured while upgrading the websocket connection")
		return
	}
}
