package stream

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type req struct {
	Accuracy *float64 `json:"accuracy"`
	Status   *int64   `json:"status" validate:"omitempty,oneof=0 1 2 3 4 5"`
	Heading  *float64 `json:"heading"`
	Lat      float64  `json:"lat" validate:"required,latitude"`
	Lon      float64  `json:"lon" validate:"required,longitude"`
}

func add(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)
	count := 1

	writer := c.K.B
	writer.Balancer = kafka.BalancerFunc(func(m kafka.Message, i ...int) int {
		return partitionNo
	})
	writer.Async = true

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	upgrader.OnMessage(func(_ *websocket.Conn, _ websocket.MessageType, b []byte) {
		var (
			data    req
			payload string
			err     error
		)

		if err = sonic.UnmarshalString(string(b), &data); err != nil {
			log.Error().Err(err).Msg("provide valid JSON data")
			return
		}
		if err = v.Struct(data); err != nil {
			log.Error().Err(err).
				Msgf(
					"body : %v\tprovided data with the websocket connection is not valid",
					data,
				)
			return
		}

		payload, err = sonic.MarshalString(blob(data))
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"data : %v\tpartition : %d\tdriver_id : %d\tfailed to marshal the payload",
					data,
					partitionNo,
					driverID,
				)
			return
		}

		writer.WriteMessages(context.Background(), kafka.Message{
			Key:   []byte(strconv.Itoa(int(driverID))),
			Value: []byte(payload),
		})

		if count < updateinterval {
			count++
		} else {
			c.R.DB.Set(r.Context(), lib.L(partitionNo), payload, redis.KeepTTL)
			count = 1
		}
	})
	upgrader.OnOpen(func(conn *websocket.Conn) {
		log.Info().
			Msgf(
				"addr : %s\tconnection opened",
				conn.RemoteAddr().String(),
			)
		done := make(chan struct{})
		closed := int32(0)

		go func() {
			ticker := time.NewTicker(heartbeat)
			defer ticker.Stop()

			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					if isClosed(&closed) {
						return
					}
					conn.WriteMessage(websocket.PingMessage, nil)
				}
			}
		}()

		conn.OnClose(func(c *websocket.Conn, err error) {
			close(done)
			atomic.StoreInt32(&closed, 1)

			if err != nil {
				log.Error().Err(err).
					Msgf(
						"addr : %s\tconnection closed with error",
						c.RemoteAddr().String(),
					)
			} else {
				log.Info().
					Msgf(
						"addr : %s\tconnection closed",
						c.RemoteAddr().String(),
					)
			}
		})
	})

	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("error occured while upgrading the websocket connection")
		return
	}
}

func blob(
	payload req,
) map[string]any {
	return map[string]any{
		"lat": payload.Lat,
		"lon": payload.Lon,
		"heading": func() float64 {
			if payload.Heading == nil {
				return 0
			}
			return *payload.Heading
		}(),
		"accuracy": func() float64 {
			if payload.Accuracy == nil {
				return -1
			}
			return *payload.Accuracy
		}(),
		"status": func() int {
			if payload.Status == nil {
				return int(lib.DefaultStatus)
			}

			return int(*payload.Status)
		}(),
	}
}
