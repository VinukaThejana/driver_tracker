package routes

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog/log"
)

// EndStream is a route used by the driver to end the stream
type EndStream struct {
	E *env.Env
	C *connections.C
}

// Method contains the method of the EndStream rotue
func (end *EndStream) Method() string {
	return http.MethodPost
}

// Path contains the HTTP path for the endStream route
func (end *EndStream) Path() string {
	return "/end"
}

// Handler contians the logic of the endStream route
func (end *EndStream) Handler(w http.ResponseWriter, r *http.Request) {
	const maxRequestBodySize = 1 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	var reqBody body
	v := validator.New()

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendJSONResponse(w, http.StatusUnsupportedMediaType, "failed to decode the request body")
		return
	}

	if err := v.Struct(reqBody); err != nil {
		log.Error().Err(err).Msg("validation error, invalid data is provided")
		sendJSONResponse(w, http.StatusBadRequest, "please provide a proper driver id")
		return
	}

	dialer, conn, err := end.C.GetKafkaConnection(end.E)
	if err != nil {
		log.Error().Err(err).Msg("failed to end kafka connection")
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
		log.Error().Err(err).Msg("failed to end the controller connection")
		sendJSONResponse(w, http.StatusInternalServerError, "something went wrong on our end")
		return
	}
	defer controllerConn.Close()

	err = controllerConn.DeleteTopics(fmt.Sprintf("%d", reqBody.DriverID))
	if err != nil {
		log.Error().Err(err).Msg("failed to delete the topic")
		sendJSONResponse(w, http.StatusInternalServerError, "failed to end the stream")
		return
	}

	sendJSONResponse(w, http.StatusOK, "ended the stream")
}
