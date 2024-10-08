package bookings

import (
	"context"
	"net/http"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

func index(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	bookingID := ""
	client := c.R.DB

	response := struct {
		Active   []string `json:"active"`
		Inactive []string `json:"inactive"`
	}{
		Active:   []string{},
		Inactive: []string{},
	}

	jobs := client.SMembers(r.Context(), e.PartitionManagerKey).Val()
	if len(jobs) == 0 {
		lib.JSONResponseWInterface(w, http.StatusOK, response)
		return
	}

	for _, job := range jobs {
		val := client.Get(r.Context(), _lib.C(job)).Val()
		if val == "" {
			bookingID = getBookingID(r.Context(), e, client, job)
			if bookingID == "" {
				continue
			}

			response.Inactive = append(response.Inactive, bookingID)
			continue
		}

		bookingID = getBookingID(r.Context(), e, client, job)
		response.Active = append(response.Active, bookingID)
	}

	lib.JSONResponseWInterface(w, http.StatusOK, response)
}

func getBookingID(ctx context.Context, e *env.Env, client *redis.Client, job string) (bookingID string) {
	val := client.Get(ctx, _lib.N(job)).Val()
	if val == "" {
		log.Warn().
			Msgf(
				"job : %s\tvalue : %s\tthe n- job has also been deleted",
				job,
				val,
			)

		err := client.SRem(ctx, e.PartitionManagerKey, job).Err()
		if err != nil {
			log.Error().Err(err).
				Msgf(
					"job : %s\tfailed to delete the job from the job manager",
					job,
				)
		}
		return ""
	}

	N := _lib.NewN()
	err := sonic.UnmarshalString(val, &N)
	if err != nil || len(N) != 2 {
		log.Error().Err(err).
			Msgf(
				"job : %s\tpayload : %v\tfailed to unmarshal the job with n-",
				job,
				val,
			)
		return ""
	}

	return N[0]
}
