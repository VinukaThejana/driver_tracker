package stream

import (
	"context"
	"io"
	"net/http"
	"strconv"

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

// add is a route that is used to add data to the stream
func addV2(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 8
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var (
		reqData struct {
			Location struct {
				Heading  *float64 `json:"heading"`
				Accuracy *float64 `json:"accuracy"`
				Status   *int64   `json:"status" validate:"omitempty,oneof=0 1 2 3 4 5"`
				Lat      float64  `json:"lat" validate:"required,latitude"`
				Lon      float64  `json:"lon" validate:"required,longitude"`
			} `json:"location" validate:"required"`
		}
		payload string
		err     error
	)

	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqData)
	if err != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			log.Error().Err(err).
				Msg("failed to read the request body")
			lib.JSONResponse(w, http.StatusRequestEntityTooLarge, errors.ErrBadRequest.Error())
			return
		}
		log.Error().Err(err).
			Msgf(
				"raw_body : %s\tfailed to read the request body",
				string(body),
			)

		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	if err = v.Struct(reqData); err != nil {
		log.Error().Err(err).
			Msgf(
				"body : %v\tfailed to validate the request body",
				reqData,
			)
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBadRequest.Error())
		return
	}
	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	blob := blob(req{
		Lat:      reqData.Location.Lat,
		Lon:      reqData.Location.Lon,
		Heading:  reqData.Location.Heading,
		Accuracy: reqData.Location.Accuracy,
		Status:   reqData.Location.Status,
	})

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
