package routes

import (
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// CreateStream is a route that is used to create a stream for the given driver
type CreateStream struct {
	E *env.Env
	C *connections.C
}

// Method contains the method of the createStream route
func (create *CreateStream) Method() string {
	return http.MethodPost
}

// Path contains the HTTP path for the createStream route
func (create *CreateStream) Path() string {
	return "/create"
}

type body struct {
	BookingID string `json:"booking_id" validate:"required,min=1"`
}

// Handler contains the bussiness logic of the createStream route
func (create *CreateStream) Handler(w http.ResponseWriter, r *http.Request) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := sonic.ConfigDefault.NewDecoder(r.Body); err != nil {
		sendJSONResponse(w, http.StatusUnsupportedMediaType, "failed to decode the request body")
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		sendJSONResponse(w, http.StatusBadRequest, "please provide a proper driver id")
		return
	}

	dialer, conn, err := create.C.GetKafkaConnection(create.E)
	if err != nil {
		log.Error().Err(err).Msg("failed to create kafka connection")
		sendJSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Error().Err(err).Msg("failed to connecto to kafka")
		sendJSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}

	controllerConn, err := dialer.Dial("tcp", net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port)))
	if err != nil {
		log.Error().Err(err).Msg("failed to create the controller connection")
		sendJSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
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
		sendJSONResponse(w, http.StatusInternalServerError, "failed to create the stream")
		return
	}

	sendJSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"sub_url": map[string]interface{}{
			"ws":   fmt.Sprintf("ws://%s/ws/view/%s", create.E.Host, reqBody.BookingID),
			"http": fmt.Sprintf("%s/view/%s", create.E.Domain, reqBody.BookingID),
		},
		"pub_url": map[string]interface{}{
			"ws":   fmt.Sprintf("ws://%s/ws/add/%s", create.E.Host, reqBody.BookingID),
			"http": fmt.Sprintf("%s/add/%s", create.E.Domain, reqBody.BookingID),
		},
	})
}
