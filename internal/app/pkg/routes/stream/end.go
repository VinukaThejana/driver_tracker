package stream

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/errors"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/lib"
	"github.com/rs/zerolog/log"
)

// End is a route that is used to end a given stream
func end(w http.ResponseWriter, r *http.Request, e *env.Env, c *connections.C) {
	id := r.Context().Value(middlewares.BookingTokenID).(string)
	bookingID := r.Context().Value(middlewares.BookingID).(string)
	partitionNo := r.Context().Value(middlewares.PartitionNo).(int)

	payload := make([]int, 2)
	err := sonic.UnmarshalString(c.R.DB.Get(r.Context(), bookingID).Val(), &payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to backup the job")
		log.Warn().Interface("payload", payload)
		return
	}

	partition := payload[0]

	pipe := c.R.DB.Pipeline()

	pipe.Del(r.Context(), id)
	pipe.Del(r.Context(), bookingID)
	pipe.Del(r.Context(), fmt.Sprint(partitionNo))
	pipe.SRem(r.Context(), e.PartitionManagerKey, partitionNo)

	_, err = pipe.Exec(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("failed to end the session")
		lib.JSONResponse(w, http.StatusInternalServerError, errors.ErrServer.Error())
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "booking_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(-time.Hour * 24),
	})
	lib.JSONResponse(w, http.StatusOK, "ended the session")

	go func() {
		ctx := context.Background()

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
		payload, err := sonic.MarshalString(messages)
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
		_, err = fmt.Fprint(w, payload)
		if err != nil {
			log.Error().Err(err).Msg("failed to write the messages to the google cloud storage")
			return
		}
	}()
}
