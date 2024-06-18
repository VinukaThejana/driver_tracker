package stream

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
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
		if errors.Is(err, io.EOF) {
			log.Error().Msg("request body is empty")
			lib.JSONResponse(w, http.StatusBadRequest, "body cannot be empty, please provide valid json")
			return
		}

		log.Error().Msg("invalid request body provided by the client")
		lib.JSONResponse(w, http.StatusBadRequest, "failed to parse data invalid json")
		return
	}

	driverID := r.Context().Value(middlewares.DriverID).(int)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)
	data["timestamp"] = time.Now().UTC().Unix()

	payload, err = sonic.MarshalString(data)
	if err != nil {
		log.Error().Err(err).Msg("error marshaling the request body")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on the server side")
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

	lib.JSONResponse(w, http.StatusOK, "added the message to the topic")
}
