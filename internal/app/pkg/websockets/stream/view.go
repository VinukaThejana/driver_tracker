package stream

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}

	client := c.R.DB

	val := client.Get(r.Context(), bookingID).Val()
	if val == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}
	payload := make([]int, 2)
	err := sonic.UnmarshalString(val, &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal the value from Redis")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	partition := payload[0]
	cKey := fmt.Sprintf("c%d", partition)

	val = client.Get(r.Context(), cKey).Val()
	if val == "" {
		log.Warn().Msg("number of connections is not present in the redis db")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	connections, err := strconv.Atoi(val)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert the maximum connections to int")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	if connections >= e.MaxConnections {
		log.Warn().
			Str("bookingID", bookingID).
			Int("connections", connections).
			Msg("maximum number of connections reached for the booking")
		http.Error(w, errors.ErrBadRequest.Error(), http.StatusTooManyRequests)
		return
	}
	client.Incr(r.Context(), cKey)

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
			go func() {
				ctx := context.TODO()

				val := client.Get(ctx, cKey).Val()
				connections, err := strconv.Atoi(val)
				if err != nil {
					log.Error().Err(err).Msg("failed to convert the connections to int")
					resetActiveConn(ctx, client, cKey)
					return
				}
				if connections <= 0 {
					resetActiveConn(ctx, client, cKey)
					return
				}

				client.Decr(ctx, cKey)
			}()

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

func resetActiveConn(ctx context.Context, client *redis.Client, key string) {
	ttl := client.TTL(ctx, key).Val()
	client.SetNX(ctx, key, 0, ttl)
}
