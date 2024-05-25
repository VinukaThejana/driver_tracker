package services

import (
	"testing"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

var (
	e         env.Env
	connector connections.C
)

func init() {
	e.Load("../../../", ".env")
	connector.InitRedis(&e)
}

func TestLog(t *testing.T) {
}
