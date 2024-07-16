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
	driverID := r.Context().Value(middlewares.DriverID).(int)

	client := c.R.DB

	val := client.Get(r.Context(), fmt.Sprint(driverID)).Val()
	if val != id {
		lib.JSONResponse(w, http.StatusUnauthorized, errors.ErrUnauthorized.Error())
		return
	}

	payload := make([]int, 3)
	err := sonic.UnmarshalString(c.R.DB.Get(r.Context(), bookingID).Val(), &payload)
	if err != nil {
		log.Error().Err(err).Interface("payload", payload).Msg("failed to backup the job")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	pipe := client.Pipeline()

	pipe.Del(r.Context(), bookingID)
	pipe.Del(r.Context(), fmt.Sprint(driverID))
	pipe.Del(r.Context(), fmt.Sprintf("n%d", partitionNo))
	pipe.Del(r.Context(), fmt.Sprintf("c%d", partitionNo))

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

	go services.GenerateLog(e, c, payload, bookingID)
}
