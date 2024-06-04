package stream

import (
	"errors"
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

// End is a route that is used to end a given stream
func end(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := sonic.ConfigDefault.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Error().Err(err).Msg("failed to decode the request body")
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
		log.Error().Err(err).Msg("failed to end kafka connection")
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
		log.Error().Err(err).Msg("failed to end the controller connection")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}
	defer controllerConn.Close()

	err = controllerConn.DeleteTopics(reqBody.BookingID)
	if err != nil {
		if errors.Is(err, kafka.UnknownTopicID) || errors.Is(err, kafka.UnknownTopicOrPartition) {
			lib.JSONResponse(w, http.StatusBadRequest, fmt.Sprintf("driver of id : %s does not have an active stream", reqBody.BookingID))
			return
		}

		log.Error().Err(err).Msg("failed to delete the topic")
		lib.JSONResponse(w, http.StatusInternalServerError, "failed to end the stream")
		return
	}

	lib.JSONResponse(w, http.StatusOK, "ended the stream")
}
