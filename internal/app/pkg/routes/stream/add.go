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

// add is a route that is used to add data to the stream
func add(w http.ResponseWriter, r *http.Request, _ *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var (
		data    map[string]interface{}
		payload string
		err     error
	)

	err = sonic.ConfigDefault.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		log.Error().Err(err).Msg("request body is not supported")
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, errors.ErrUnsuportedMedia.Error())
		return
	}

	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)
	data["timestamp"] = time.Now().UTC().Unix()

	payload, err = sonic.MarshalString(data)
	if err != nil {
		log.Error().Err(err).Msg("error marshaling the request body")
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
