package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
)

// Log is used as a log drain to log messages
func Log(c *connections.C, log interface{}) {
	now := time.Now()
	timestamp := fmt.Sprintf("[%d|%s|%d : %d:%d]", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute())
	jsonB, err := json.Marshal(log)
	if err != nil {
		fmt.Println(log)
		logger.Error(err)
	}

	err = c.R.Log.SetEx(context.Background(), timestamp, string(jsonB), time.Second*24*60*60).Err()
	if err != nil {
		fmt.Println(err)
		logger.Error(err)
	}
}
