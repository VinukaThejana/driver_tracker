package routes

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// Add a message to given Kafka topic
type Add struct {
	E *env.Env
	C *connections.C
}

// Method HTTP method used by the Add route
func (add *Add) Method() string {
	return http.MethodPost
}

// Path is the route path used by the Add route
func (add *Add) Path() string {
	return "/add/{topic}"
}

// Handler is the bussiness logic of the Add route
func (add *Add) Handler(w http.ResponseWriter, r *http.Request) {
	const maxRequestBodySize = 1 << 20

	contentType := r.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "application/json") {
		log.Error().Str("Content-Type", contentType).Msg("invalid content type provided")
		sendJSONResponse(w, http.StatusBadRequest, "only content of type application/json can be sent")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	topic := chi.URLParam(r, "topic")
	if topic == "" {
		log.Error().Msg("topic is not provided")
		sendJSONResponse(w, http.StatusBadRequest, "topic is not provided")
		return
	}

	var data map[string]interface{}

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Error().Msg("request body is empty")
			sendJSONResponse(w, http.StatusBadRequest, "body cannot be empty, please provide valid json")
			return
		}

		log.Error().Msg("invalid request body provided by the client")
		sendJSONResponse(w, http.StatusBadRequest, "failed to parse data invalid json")
		return
	}
	data["timestamp"] = time.Now().UTC().Unix()

	dataBytes, err := json.Marshal(data)
	if err != nil {
		log.Error().Err(err).Msg("error marshaling the request body")
		sendJSONResponse(w, http.StatusInternalServerError, "something went wrong on the server side")
		return
	}
	dataStr := string(dataBytes)

	add.C.KafkaWriteToTopicWithHTTP(add.E, topic, dataStr)

	sendJSONResponse(w, http.StatusOK, "added the message to the topic")
}
