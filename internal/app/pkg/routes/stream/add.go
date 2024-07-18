package stream

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type req struct {
	Accuracy      *float64 `json:"accuracy"`
	SpeedAccuracy *float64 `json:"speed_accuracy"`
	Heading       *float64 `json:"heading"`
	Lat           float64  `json:"lat" validate:"required,latitude"`
	Lon           float64  `json:"lon" validate:"required,longitude"`
}

// add is a route that is used to add data to the stream
func add(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 7
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var (
		data    req
		payload string
		err     error
	)

	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Msg("request body is not supported")
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	if err = v.Struct(data); err != nil {
		log.Error().Err(err).Interface("body", data).Msg("failed to validate the request body")
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBadRequest.Error())
		return
	}
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	blob := blob(data)

	payload, err = sonic.MarshalString(blob)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the payload")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	log.Debug().
		Int("partition", partitionNo).
		Int("driver_id", driverID).
		Msg(payload + " " + fmt.Sprint(partitionNo) + " " + fmt.Sprint(driverID))
	go func(payload string) {
		err = c.R.DB.Set(r.Context(), _lib.L(partitionNo), payload, redis.KeepTTL).Err()
		if err != nil {
			log.Error().Err(err).Str("payload", payload).Msg("failed to set the live location")
		}
	}(payload)

	writer := c.K.B
	writer.Balancer = kafka.BalancerFunc(func(m kafka.Message, i ...int) int {
		return partitionNo
	})
	writer.Async = true
	writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(strconv.Itoa(int(driverID))),
		Value: []byte(payload),
	})

	log.Info().
		Int("partition", partitionNo).
		Int("driver_id", driverID).
		Msg("recorded the location ... ")
	lib.JSONResponse(w, http.StatusOK, "added")
}

func blob(payload req) map[string]any {
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
		"speed_accuracy": func() float64 {
			if payload.SpeedAccuracy == nil {
				return -1
			}
			return *payload.SpeedAccuracy
		}(),
		"timestamp": time.Now().UTC().Unix(),
	}
}
