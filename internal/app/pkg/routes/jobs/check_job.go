package jobs

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/bytedance/sonic"
	_lib "github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
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

			val := client.Get(r.Context(), _lib.C(job)).Val()
			if val != "" {
				return
			}
			client.Del(r.Context(), _lib.L(job))

			val = client.Get(r.Context(), _lib.N(job)).Val()
			if val == "" {
				partition, err := strconv.Atoi(job)
				msg := fmt.Sprintf(
					"job : %s\tredis-value : %s\tthe job cannot be found with n- partitions",
					job,
					val,
				)
				if err == nil {
					log.Warn().Msg(msg)
					_lib.Free(r.Context(), client, e.PartitionManagerKey, partition)
					return
				}

				log.Error().Msg(msg)
				return
			}

			N := _lib.NewN()
			err := sonic.UnmarshalString(val, &N)
			if err != nil {
				log.Error().Err(err).
					Msgf(
						"job : %s\tredis-value : %s\tcannot unmarshal the payload",
						job,
						val,
					)
				return
			}
			startOffset, err := strconv.Atoi(N[_lib.NLastOffset])
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

			pipe := client.Pipeline()

			pipe.Del(r.Context(), _lib.N(job))
			pipe.Del(r.Context(), _lib.L(job))
			pipe.Del(r.Context(), _lib.C(job))

			_, err = pipe.Exec(r.Context())
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
				N[_lib.NBookingID],
				partition,
				int64(startOffset),
			)
		}(job)
	}

	fmt.Fprint(w, "okay")
}
