package stream

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/types"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

func add(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	bookingTokenID := chi.URLParam(r, "booking_token_id")
	if bookingTokenID == "" {
		http.Error(w, errors.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}
	driverID, err := strconv.Atoi(chi.URLParam(r, "driver_id"))
	if err != nil {
		log.Error().Err(err).Msgf(
			"booking_token_id : %s\tfailed to convert the driver ID to integer",
			bookingTokenID,
		)
		http.Error(w, errors.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}
	partitionNo, err := strconv.Atoi(chi.URLParam(r, "partition"))
	if err != nil {
		log.Error().Err(err).Msgf(
			"booking_token_id : %s\tdriver_id : %d\tfailed to convert the partition number to integer",
			bookingTokenID,
			driverID,
		)
		http.Error(w, errors.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	basicDebugMsg := fmt.Sprintf(
		"booking_token_id : %s\tdriver_id : %d\tpartition : %d",
		bookingTokenID,
		driverID,
		partitionNo,
	)

	client := c.R.DB
	val := client.Get(r.Context(), fmt.Sprint(driverID)).Val()
	if val == "" {
		log.Error().Msgf(
			"%s\tvalue obtained for the driver ID is empty",
			basicDebugMsg,
		)
		http.Error(w, errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}
	DriverID := _lib.NewDriverID()
	err = sonic.UnmarshalString(val, &DriverID)
	if err != nil {
		log.Error().Err(err).Msgf(
			"%s\tfailed to unmarshal the driver ID",
			basicDebugMsg,
		)
		http.Error(w, errors.ErrBadRequest.Error(), http.StatusBadRequest)
		return
	}

	if DriverID[_lib.DriverIDDriverToken] != bookingTokenID {
		log.Error().Msgf(
			"%s\tbooking_token_id_in_redis : %s\tunmatched driver tokens",
			basicDebugMsg,
			DriverID[_lib.DriverIDDriverToken],
		)
		http.Error(w, errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

	if DriverID[_lib.DriverIDPartitionNo] != fmt.Sprint(partitionNo) {
		log.Error().Msgf(
			"%s\tpartition_number mismatch",
			basicDebugMsg,
		)
		http.Error(w, errors.ErrUnauthorized.Error(), http.StatusUnauthorized)
		return
	}

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
			data struct {
				Location types.LocationUpdate `json:"location" validate:"required"`
			}
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

		payload, err = sonic.MarshalString(data.Location.GetBlob())
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
			client.Set(r.Context(), _lib.L(partitionNo), payload, redis.KeepTTL)
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

	_, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().
			Err(err).
			Msg("error occured while upgrading the websocket connection")
		return
	}
}
