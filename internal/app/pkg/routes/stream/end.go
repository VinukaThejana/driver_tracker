package stream

import (
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

// End is a route that is used to end a given stream
func end(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	id := r.Context().Value(middlewares.BookingTokenID).(string)
	bookingID := r.Context().Value(middlewares.BookingID).(string)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	pipe := c.R.DB.Pipeline()

	pipe.Del(r.Context(), id)
	pipe.Del(r.Context(), bookingID)
	pipe.SRem(r.Context(), e.PartitionManagerKey, partitionNo)

	_, err := pipe.Exec(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to end the session")
		lib.JSONResponse(w, http.StatusInternalServerError, "something went wrong, please try again")
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
}
