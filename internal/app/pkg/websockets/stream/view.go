package stream

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	errs "github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

const (
	heartbeat = 5 * time.Second
	pending   = 2 * time.Second
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
		http.Error(w, errs.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	partition := payload[0]
	cKey := fmt.Sprintf("c%d", partition)

	val = client.Get(r.Context(), cKey).Val()
	if val == "" {
		log.Warn().Msg("number of connections is not present in the redis db")
		http.Error(w, errs.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	connections, err := strconv.Atoi(val)
	if err != nil {
		log.Error().Err(err).Msg("failed to convert the maximum connections to int")
		http.Error(w, errs.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	if connections >= e.MaxConnections {
		log.Warn().
			Str("bookingID", bookingID).
			Int("connections", connections).
			Msg("maximum number of connections reached for the booking")
		http.Error(w, errs.ErrBadRequest.Error(), http.StatusTooManyRequests)
		return
	}
	client.Incr(r.Context(), cKey)

	location := client.Get(r.Context(), fmt.Sprintf("l%d", partition)).Val()
	if location == "" {
		log.Error().Msg("failed to get the last location from redis")
		http.Error(w, errs.ErrServer.Error(), http.StatusInternalServerError)
		return
	}

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Info().Str("addr", conn.RemoteAddr().String()).Msg("connection opened")
		done := make(chan struct{})
		closed := int32(0)

		go func() {
			ticker := time.NewTicker(heartbeat)
			reader := c.KafkaReader(e, e.Topic, partition, kafka.LastOffset)

			defer func() {
				reader.Close()
				ticker.Stop()
			}()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if isClosed(&closed) {
						return
					}

					log.Info().Msg("heartbeat ... ")
					conn.WriteMessage(websocket.PingMessage, nil)
				default:
					ctx, cancel = context.WithTimeout(r.Context(), pending)
					message, err := reader.ReadMessage(ctx)
					cancel()

					if err != nil {
						if errors.Is(err, context.DeadlineExceeded) {
							if isClosed(&closed) {
								continue
							}

							if err := conn.WriteMessage(websocket.TextMessage, []byte(location)); err != nil {
								log.Error().Err(err).Msg("error sending data to the websocket client")
							}
							continue
						}
						log.Error().Err(err).Msg("error reading message from kafka")
						continue
					}

					if len(message.Value) == 0 || isClosed(&closed) {
						continue
					}

					location = string(message.Value)
					if err := conn.WriteMessage(websocket.TextMessage, message.Value); err != nil {
						log.Error().Err(err).Msg("error sending data to the websocket client")
						continue
					}
				}
			}
		}()

		conn.OnClose(func(c *websocket.Conn, err error) {
			close(done)
			atomic.StoreInt32(&closed, 1)

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

func isClosed(closed *int32) bool {
	return atomic.LoadInt32(closed) == 1
}

func resetActiveConn(
	ctx context.Context,
	client *redis.Client,
	key string,
) {
	ttl := client.TTL(ctx, key).Val()
	client.SetNX(ctx, key, 0, ttl)
}
