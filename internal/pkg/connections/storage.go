package connections

import (
	"context"
	"encoding/base64"

	"cloud.google.com/go/storage"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"google.golang.org/api/option"
)

// InitStorage is a function that is used to initialize the Google cloud storage
func (c *C) InitStorage(e *env.Env) {
	key, err := base64.StdEncoding.DecodeString(e.GcloudAPIKey)
	if err != nil {
		lib.LogFatal(err)
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(key))
	if err != nil {
		lib.LogFatal(err)
	}

	c.S = client
}
