package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

// Log is used as a log drain to log messages
func Log(c *connections.C, e *env.Env, log interface{}) {
	now := time.Now()
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		timestamp := fmt.Sprintf("[%d|%s|%d : %d:%d]", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
		jsonB, err := json.Marshal(log)
		if err != nil {
			fmt.Println(log)
			logger.Error(err)
		}

		err = c.R.Log.Set(context.Background(), timestamp, string(jsonB), 0).Err()
		if err != nil {
			fmt.Println(err)
			logger.Error(err)
		}
	}()

	go func() {
		defer wg.Done()

		type M map[string]interface{}
		payload, err := json.Marshal(M{
			"date":      fmt.Sprintf("%d %s %d", now.Day(), now.Month(), now.Year()),
			"time":      fmt.Sprintf("%d : %d", now.Hour(), now.Minute()),
			"timestamp": now.UTC().Unix(),
			"data":      log,
		})
		if err != nil {
			fmt.Println(err)
			logger.Error(err)
		}

		c.KafkaWriteToTopic(e, "log", string(payload))
	}()

	wg.Wait()
}
