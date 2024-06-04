package stream

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

type body struct {
	BookingID string `json:"booking_id" validate:"required,min=1"`
}

// Create is a route that is used to create a new stream for the given booking id
func create(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		lib.JSONResponse(w, http.StatusUnsupportedMediaType, "failed to decode the request body")
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		lib.JSONResponse(w, http.StatusBadRequest, "please provide a proper driver id")
		return
	}

	dialer, conn, err := c.GetKafkaConnection(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to create kafka connection")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Error().Err(err).Msg("failed to connecto to kafka")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}

	controllerConn, err := dialer.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the controller connection")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}
	defer controllerConn.Close()

	err = controllerConn.CreateTopics(kafka.TopicConfig{
		Topic:             reqBody.BookingID,
		NumPartitions:     1,
		ReplicationFactor: 1,
	})
	if err != nil {
		log.Error().Err(err).Msg("failed to create the topic")
		lib.JSONResponse(w, http.StatusInternalServerError, "failed to create the stream")
		return
	}

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"sub_url": map[string]interface{}{
			"ws": fmt.Sprintf("ws://%s/ws/view/%s", e.Host, reqBody.BookingID),
		},
		"pub_url": map[string]interface{}{
			"ws":   fmt.Sprintf("ws://%s/ws/add/%s", e.Host, reqBody.BookingID),
			"http": fmt.Sprintf("%s/add/%s", e.Domain, reqBody.BookingID),
		},
	})
}
