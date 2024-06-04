package stream

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

// add is a route that is used to add data to the stream
func add(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	const maxRequestBodySize = 1 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
	defer r.Body.Close()

	topic := chi.URLParam(r, "topic")
	if topic == "" {
		log.Error().Msg("topic is not provided")
		lib.JSONResponse(w, http.StatusBadRequest, "topic is not provided")
		return
	}

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
	data["timestamp"] = time.Now().UTC().Unix()

	payload, err = sonic.MarshalString(data)
	if err != nil {
		log.Error().Err(err).Msg("error marshaling the request body")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong on the server side")
		return
	}

	err = c.KafkaWriteToTopicWithHTTP(e, topic, payload)
	if err != nil {
		log.Error().Err(err).Msg("error sending the message to the topic")
		if errors.Is(err, connections.ErrKafkaNoTopic) {
			lib.JSONResponse(w, http.StatusBadRequest, err.Error())
			return
		}
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	lib.JSONResponse(w, http.StatusOK, "added the message to the topic")
}
