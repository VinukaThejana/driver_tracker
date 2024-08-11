// Package stream contains websockets that are related to the stream
package stream

import (
	"net/http"
	"sync/atomic"
	"time"

	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/lib"
	"github.com/flitlabs/spotoncars_stream/internal/app/pkg/middlewares"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/connections"
	"github.com/flitlabs/spotoncars_stream/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

const (
	// updateinterval is the update frequency of redis
	updateinterval = 5
	// heartbeat is the frequency to send a ping to the websocket client
	heartbeat = 5 * time.Second
	// pending is the deadline to keep wating for the kafka reader
	pending = 2 * time.Second
)

var (
	v = validator.New()
	h = lib.WrapHandler
	m = lib.WrapMiddleware
)

// WebSocket contains all the websockets that are related to the stream
func WebSocket(e *env.Env, c *connections.C) http.Handler {
	r := chi.NewRouter()

	r.Get("/view/{booking_id}", h(view, e, c))

	r.Route("/create", func(r chi.Router) {
		r.Use(m(func(h http.Handler, e *env.Env, c *connections.C) http.Handler {
			return middlewares.IsBookingTokenValid(h, e, c, true)
		}, e, c))
		r.Get("/", h(create, e, c))
	})

	r.Get("/add/{booking_token_id}/{driver_id}/{partition}", h(add, e, c))

	return r
}

// isClosed is used to check wether the websocket connection is closed in a concurrent
// safe manner
func isClosed(closed *int32) bool {
	return atomic.LoadInt32(closed) == 1
}
