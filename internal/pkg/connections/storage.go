package connections

import (
	"context"
	"encoding/base64"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"google.golang.org/api/option"
)

// InitStorage is a function that is used to initialize the Google cloud storage
func (c *C) InitStorage(e *env.Env) error {
	fmt.Println(e.GcloudAPIKey)
	key, err := base64.StdEncoding.DecodeString(e.GcloudAPIKey)
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
