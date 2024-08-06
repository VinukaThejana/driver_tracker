package services

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// GenerateLog is a function that is used to save the booking history of a particular booking
func GenerateLog(
	e *env.Env,
	c *connections.C,
	bookingID string,
	partition int,
	startOffset int64,
) {
	ctx := context.Background()
	defer free(ctx, c.R.DB, e.PartitionManagerKey, partition)

	Revalidate(e, []Paths{
		Dashboard,
	})

	endOffset, err := c.GetLastOffset(ctx, e, e.Topic, partition)

	if startOffset >= endOffset {
		log.Warn().Msg("no messages in the given partition")
		return
	}
	if err != nil {
		log.Error().Err(err).Msg("failed to get the last offset")
		log.Warn().Msgf("partition : %d", partition)
		return
	}

	defer func() {
		log.Warn().
			Msgf(
				"start : %d\tend : %d",
				int(startOffset),
				int(endOffset),
			)
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
		log.Error().Err(err).
			Msgf(
				"start : %d\tend : %d\tfailed to initialize the storage client",
				int(startOffset),
				int(endOffset),
			)
		return
	}
	defer c.S.Close()

	bucket := c.S.Bucket(e.BucketName)
	object := bucket.Object(bookingID)

	w := object.NewWriter(ctx)
	defer w.Close()
	_, err = fmt.Fprint(w, data)
	if err != nil {
		log.Error().
			Err(err).
			Msg("failed to write the messages to the google cloud storage")
		return
	}
}

// deallocate the used partition for upcomming jobs
func free(ctx context.Context, client *redis.Client, key string, partition int) {
	err := client.SRem(ctx, key, partition).Err()
	if err != nil {
		log.Error().Err(err).
			Msgf(
				"partition : %d\tfailed to remove the partition",
				partition,
			)
	}
}
