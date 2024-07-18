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
					Str("job", job).
					Str("redis-value", val).
					Msg("the job cannot be found with n- partitions")
				return
			}

			var payload []string
			err := sonic.UnmarshalString(val, &payload)
			if err != nil {
				log.Error().
					Err(err).
					Str("job", job).
					Str("redis-value", val).
					Msg("cannot unmarshal the payload")
				return
			}
			if len(payload) != 2 {
				log.Error().
					Err(
						fmt.
							Errorf("the payload did not contains the booking ID and the offset")).
					Str("job", job).
					Str("redis-value", val).
					Interface("payload", payload).
					Msg("the payload in n- partition is not equal to two")
				return
			}
			startOffset, err := strconv.Atoi(payload[1])
			if err != nil {
				log.Error().
					Err(err).
					Str("job", job).
					Str("redis-value", val).
					Msg("failed to convert the offset to int")
				return
			}
			partition, err := strconv.Atoi(job)
			if err != nil {
				log.Error().
					Err(err).
					Str("job", job).
					Str("redis-value", val).
					Msg("failed to convert the job to integer")
				return
			}

			err = client.Del(r.Context(), lib.N(job)).Err()
			if err != nil {
				log.Error().
					Err(err).
					Str("job", job).
					Str("redis-value", val).
					Msg("failed to delete the key in redis")
			}

			go services.GenerateLog(e, c, []int{partition, startOffset}, payload[0])
		}(job)
	}

	fmt.Fprint(w, "okay")
}
