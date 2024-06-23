package connections

import (
	"context"

	"cloud.google.com/go/storage"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"google.golang.org/api/option"
)

// InitStorage is a function that is used to initialize the Google cloud storage
func (c *C) InitStorage(e *env.Env) error {
	key, err := lib.Base64URLDecode(e.GcloudAPIKey)
	if err != nil {
		return err
	}

	client, err := storage.NewClient(context.Background(), option.WithCredentialsJSON(key))
	if err != nil {
		return err
	}

	c.S = client
	return nil
}
