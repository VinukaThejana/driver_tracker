package jobs

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/services"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/rs/zerolog/log"
)

func checkJob(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	var wg sync.WaitGroup

	client := c.R.DB
	jobs := client.SMembers(r.Context(), e.PartitionManagerKey).Val()
	wg.Add(len(jobs))

	for _, job := range jobs {
		go func(job string) {
			defer wg.Done()

			val := client.Get(r.Context(), lib.C(job)).Val()
			if val != "" {
				return
			}

			val = client.Get(r.Context(), lib.N(job)).Val()
			if val == "" {
				log.Warn().
					Msgf(
						"job : %s\tredis-value : %s\tthe job cannot be found with n- partitions",
						job,
						val,
					)
				return
			}

			var payload []string
			err := sonic.UnmarshalString(val, &payload)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"job : %s\tredis-value : %s\tcannot unmarshal the payload",
						job,
						val,
					)
				return
			}
			if len(payload) != 2 {
				log.Error().Err(fmt.Errorf("the payload did not contains the booking ID and the offset")).
					Msgf(
						"job : %s\tredis-value : %s\tpayload : %v\tthe payload in n- partition is not equal to two",
						job,
						val,
						payload,
					)
				return
			}
			startOffset, err := strconv.Atoi(payload[lib.NLastOffset])
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"job : %s\tredis-value : %s\tfailed to convert the offset to int",
						job,
						val,
					)
				return
			}
			partition, err := strconv.Atoi(job)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"job : %s\tredis-value : %s\tfailed to convert the job to integer",
						job,
						val,
					)
				return
			}

			err = client.Del(r.Context(), lib.N(job)).Err()
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"job : %s\tredis-value : %s\tfailed to delete the key in redis",
						job,
						val,
					)
			}

			go services.GenerateLog(
				e,
				c,
				payload[lib.NBookingID],
				partition,
				int64(startOffset),
			)
		}(job)
	}

	fmt.Fprint(w, "okay")
}
