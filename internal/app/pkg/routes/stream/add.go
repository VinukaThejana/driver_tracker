package stream

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
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

	payload, err = sonic.MarshalString(map[string]interface{}{
		"lat": data.Lat,
		"lon": data.Lon,
		"heading": func() float64 {
			if data.Heading == nil {
				return 0
			}
			return *data.Heading
		}(),
		"accuracy": func() float64 {
			if data.Accuracy == nil {
				return -1
			}
			return *data.Accuracy
		}(),
		"speed_accuracy": func() float64 {
			if data.SpeedAccuracy == nil {
				return -1
			}
			return *data.SpeedAccuracy
		}(),
		"timestamp": time.Now().UTC().Unix(),
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the payload")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	writer := c.K.B
	writer.Balancer = kafka.BalancerFunc(func(m kafka.Message, i ...int) int {
		return partitionNo
	})
	writer.Async = true
	writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(strconv.Itoa(int(driverID))),
		Value: []byte(payload),
	})

	lib.JSONResponse(w, http.StatusOK, "added the message to the topic")
}
