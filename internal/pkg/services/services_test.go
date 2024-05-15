package services

import (
	"testing"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/jaswdr/faker/v2"
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
	fake := faker.New()

	Log(&connector, fake.Json().Faker.Address())
	Log(&connector, fake.Address().Country())
}
