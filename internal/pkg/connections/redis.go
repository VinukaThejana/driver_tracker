package connections

import (
	"context"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/redis/go-redis/v9"
)

// Redis contains all Redis connections
type Redis struct {
	DB *redis.Client
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
