package logs

import (
	"fmt"
	"net/http"

	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func delete(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	err := c.InitStorage(e)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize the google cloud storage client")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	err = object.Delete(r.Context())
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"booking_id : %s\tfailed to delete the object possibly the booking id is not valid",
				bookingID,
			)
		lib.JSONResponse(w, http.StatusBadRequest, errors.ErrBookingIDNotValid.Error())
		return
	}

	lib.JSONResponse(w, http.StatusOK, fmt.Sprintf("deleted the location history of the booking id : %s", bookingID))
}
