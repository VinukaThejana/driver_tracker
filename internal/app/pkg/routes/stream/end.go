package stream

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
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
	bookingID := r.Context().Value(middlewares.BookingID).(string)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)
	driverID := r.Context().Value(middlewares.DriverID).(int)

	client := c.R.DB

	BookingID := _lib.NewBookingID()
	err := sonic.UnmarshalString(c.R.DB.Get(r.Context(), bookingID).Val(), &BookingID)
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"payload : %v\tfailed to backup the job",
				BookingID,
			)
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	err = _lib.DelBooking(r.Context(), client, driverID, bookingID, partitionNo)
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

	go services.GenerateLog(
		e,
		c,
		bookingID,
		BookingID[_lib.BookingIDPartitionNo],
		int64(BookingID[_lib.BookingIDLastOffset]),
	)
}
