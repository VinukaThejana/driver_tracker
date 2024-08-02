package stream

import (
	"context"
	"io"
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

type reqDev struct {
	Heading  *float64 `json:"heading"`
	Status   *int64   `json:"status" validate:"omitempty,oneof=0 1 2 3 4 5"`
	Location struct {
		Accuracy *float64 `json:"accuracy"`
		Lat      float64  `json:"lat" validate:"required,latitude"`
		Lon      float64  `json:"lon" validate:"required,longitude"`
	} `json:"location"`
}

// add is a route that is used to add data to the stream
func addDev(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 7
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var (
		data    reqDev
		payload string
		err     error
	)

	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).
				Msg("failed to read the request body")
		} else {
			log.Error().Err(err).
				Msgf(
					"raw_body : %s\tfailed to read the request body",
					string(body),
				)
		}

		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	if err = v.Struct(data); err != nil {
		log.Error().Err(err).
			Msgf(
				"body : %v\tfailed to validate the request body",
				data,
			)
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBadRequest.Error())
		return
	}
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	blob := blobDev(data)

	payload, err = sonic.MarshalString(blob)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the payload")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	go func(payload string) {
		err = c.R.DB.Set(r.Context(), _lib.L(partitionNo), payload, redis.KeepTTL).Err()
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"payload : %s\tfailed to set the live location",
					payload,
				)
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
		Msgf(
			"partition : %d\tdriver_id : %d\trecorded the live location ... ",
			partitionNo,
			driverID,
		)
	lib.JSONResponse(w, http.StatusOK, "added")
}

func blobDev(payload reqDev) map[string]any {
	return map[string]any{
		"lat": payload.Location.Lat,
		"lon": payload.Location.Lon,
		"heading": func() float64 {
			if payload.Heading == nil {
				return 0
			}
			return *payload.Heading
		}(),
		"accuracy": func() float64 {
			if payload.Location.Accuracy == nil {
				return -1
			}
			return *payload.Location.Accuracy
		}(),
		"status": func() int {
			if payload.Status == nil {
				return int(_lib.DefaultStatus)
			}
			return int(*payload.Status)
		}(),
		"timestamp": time.Now().UTC().Unix(),
	}
}
