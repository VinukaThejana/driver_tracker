package jobs

import (
	"fmt"
	"net/http"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
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
	payload := make([]int, 3)
	err := sonic.UnmarshalString(val, &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to unmarshal the value from Redis")
		http.Error(w, errors.ErrServer.Error(), http.StatusInternalServerError)
		return
	}
	partitionNo := payload[0]
	driverID := payload[2]

	pipe := client.Pipeline()

	pipe.Del(r.Context(), bookingID)
	pipe.Del(r.Context(), fmt.Sprint(driverID))
	pipe.Del(r.Context(), fmt.Sprintf("n%d", partitionNo))
	pipe.Del(r.Context(), fmt.Sprintf("c%d", partitionNo))
	pipe.SRem(r.Context(), e.PartitionManagerKey, partitionNo)

	_, err = pipe.Exec(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to peform redis actions")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	lib.JSONResponse(w, http.StatusOK, "removed the current booking from redis")

	go services.GenerateLog(e, c, payload, bookingID)
}
