// Package services is a package that is used to manage various services
package services

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/rs/zerolog/log"
)

// SaveBooking is a function that is used to save the booking history of a particular booking
func SaveBooking(e *env.Env, c *connections.C, payload []int, bookingID string) {
	ctx := context.Background()
	partition := payload[0]

	startOffset := int64(payload[1])
	endOffset, err := c.GetLastOffset(ctx, e, e.Topic, partition)

	if startOffset >= endOffset {
		log.Warn().Msg("no messages in the given partition")
		return
	}

	if err != nil {
		log.Error().Err(err).Msg("failed to get the last offset")
		log.Warn().Interface("payload", payload)
		return
	}
	defer func() {
		log.Warn().Int("start", int(startOffset)).Int("end", int(endOffset))
	}()

	messages, err := c.GetLastNMessages(ctx, e, startOffset, endOffset, e.Topic, partition)
	if err != nil {
		log.Error().Err(err).Msg("failed to get the last messages")
		return
	}
	data, err := sonic.MarshalString(messages)
	if err != nil {
		log.Error().Err(err).Msg("failed to marshal the messages")
		return
	}

	err = c.InitStorage(e)
	if err != nil {
		log.Error().Err(err).Int("start", int(startOffset)).Int("end", int(endOffset)).Msg("failed to initialize the storage client")
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	w := object.NewWriter(ctx)
	defer w.Close()
	_, err = fmt.Fprint(w, data)
	if err != nil {
		log.Error().Err(err).Msg("failed to write the messages to the google cloud storage")
		return
	}
}
