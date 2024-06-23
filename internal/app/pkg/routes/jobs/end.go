package jobs

import (
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
)

func end(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := chi.URLParam(r, "booking_id")
	if bookingID == "" {
		lib.JSONResponse(w, http.StatusBadRequest, "please provide a valid booking id")
		return
	}

	client := c.R.DB

	val := client.Get(r.Context(), bookingID).Val()
	if val == "" {
		lib.JSONResponse(w, http.StatusBadRequest, "provided booking id is not valid")
		return
	}
	payload := make([]int, 2)
	err := sonic.UnmarshalString(val, &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal the value from Redis")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	partitionNo := payload[0]

	driverTokenID := client.Get(r.Context(), fmt.Sprint(partitionNo)).Val()
	if driverTokenID == "" {
		lib.JSONResponse(w, http.StatusUnauthorized, errors.ErrUnauthorized.Error())
		return
	}

	pipe := client.Pipeline()

	pipe.Del(r.Context(), bookingID)
	pipe.Del(r.Context(), driverTokenID)
	pipe.Del(r.Context(), fmt.Sprint(partitionNo))
	pipe.SRem(r.Context(), e.PartitionManagerKey, partitionNo)

	_, err = pipe.Exec(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to peform redis actions")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	lib.JSONResponse(w, http.StatusOK, "removed the current booking from redis")
}
