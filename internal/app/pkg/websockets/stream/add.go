package stream

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func add(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

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

		go func() {
			writer := c.K.B
			writer.Balancer = kafka.BalancerFunc(func(m kafka.Message, i ...int) int {
				return partitionNo
			})
			writer.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(strconv.Itoa(int(driverID))),
				Value: []byte(payload),
			})
		}()
	})
	upgrader.OnOpen(func(c *websocket.Conn) {
		log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection opened")
		c.OnClose(func(c *websocket.Conn, err error) {
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
