package stream

import (
	"net/http"
	"strconv"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}
	val := c.R.DB.Get(r.Context(), bookingID).Val()
	if val == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}
	partition, err := strconv.Atoi(val)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert the partition to integer")
		http.Error(w, "something went wrong, please try again later", http.StatusInternalServerError)
		return
	}

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Info().Str("addr", conn.RemoteAddr().String()).Msg("connection opened")
		done := make(chan struct{})

		go func() {
			reader := c.KafkaReader(e, e.Topic, partition, kafka.LastOffset)
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

	_, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("error occured while upgrading the websocket connection")
		return
	}
}
