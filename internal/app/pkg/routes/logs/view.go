package logs

import (
	"io"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/errors"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func view(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
	}

	err := c.InitStorage(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize the storage client")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	reader, err := object.NewReader(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize the reader")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		log.Error().Err(err).Msg("failed to read the data")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	if data == nil {
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	var payload any
	err = sonic.Unmarshal(data, &payload)
	if err != nil {
		log.Error().Err(err).Interface("payload", payload).Msg("failed to marshal the data from the payload")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	lib.JSONResponseWInterface(w, http.StatusOK, map[string]interface{}{
		"cordinates": payload,
	})
}
