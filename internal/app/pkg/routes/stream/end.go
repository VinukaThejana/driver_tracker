package stream

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

// End is a route that is used to end a given stream
func end(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	id := r.Context().Value(middlewares.BookingTokenID).(string)
	bookingID := r.Context().Value(middlewares.BookingID).(string)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	payload := make([]int, 2)
	err := sonic.UnmarshalString(c.R.DB.Get(r.Context(), bookingID).Val(), &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to backup the job")
		log.Warn().Interface("payload", payload)
		return
	}

	pipe := c.R.DB.Pipeline()

	pipe.Del(r.Context(), id)
	pipe.Del(r.Context(), bookingID)
	pipe.Del(r.Context(), fmt.Sprint(partitionNo))
	pipe.SRem(r.Context(), e.PartitionManagerKey, partitionNo)

	_, err = pipe.Exec(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to end the session")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "booking_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-time.Hour * 24),
	})
	lib.JSONResponse(w, http.StatusOK, "ended the session")

	go services.SaveBooking(e, c, payload, bookingID)
}
