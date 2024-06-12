package connections

import (
	"context"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/redis/go-redis/v9"
)

// Redis contains all Redis connections
type Redis struct {
	DB *redis.Client
}

// AcquireLock is a function that is used to lock redis keys until an operation is finished modifiying it
func (r *Redis) AcquireLock(ctx context.Context, client *redis.Client, lockKey, lockValue string, lockDuration, waitDuration time.Duration) (bool, error) {
	for {
		acquired, err := client.SetNX(ctx, lockKey, lockValue, lockDuration).Result()
		if err != nil {
			return false, err
		}

		if acquired {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			time.Sleep(waitDuration)
		}
	}
}

// ReleaseLock is a function that is used to release the added lock
func (r *Redis) ReleaseLock(client *redis.Client, lockKey string) {
	client.Del(context.Background(), lockKey)
}

// InitRedis is a function that is used to intialize redis databases
func (c *C) InitRedis(e *env.Env) {
	c.R = &Redis{
		DB: connect(e.RedisDBURL),
	}
}

func connect(redisURL string) *redis.Client {
	opt, err := redis.ParseURL(redisURL)
	lib.LogFatal(err)

	r := redis.NewClient(opt)
	lib.LogFatal(r.Ping(context.Background()).Err())

	return r
}
