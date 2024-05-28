package websockets

import (
	"context"
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/rs/zerolog/log"
	"github.com/segmentio/kafka-go"
)

// View contains the view websocket that is used to subscribe to a specific kafka topic
type View struct {
	E *env.Env
	C *connections.C
}

// Method of the view websocket
func (view *View) Method() string {
	return http.MethodGet
}

// Path is the path of the view websocket
func (view *View) Path() string {
	return "/view/{topic}"
}

// Handler is the bussiness logic of the view websocket
func (view *View) Handler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3600)
	defer cancel()

	topic := chi.URLParam(r, "topic")
	if topic == "" {
		log.Error().Msg("topic is not provided")
		return
	}

	reader := view.C.KafkaReader(view.E, topic, kafka.LastOffset)
	defer reader.Close()

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	upgrader.OnOpen(func(c *websocket.Conn) {
		log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection opened")
		for {
			message, _ := reader.ReadMessage(ctx)
			log.Error().Interface("message", message).Msg("The message is given")
			payload, err := lib.ToStr(string(message.Value))
			if err != nil {
				log.Error().Err(err).Str("value", string(message.Value)).Msg("Error occured when serialization and deserialization")
				return
			}

			if err := c.WriteMessage(websocket.TextMessage, payload); err != nil {
				log.Error().Err(err).Msg("error sending data to the websocket client")
				break
			}
		}

		c.Close()
	})

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("error occured while upgrading the websocket connection")
		return
	}

	conn.OnClose(func(c *websocket.Conn, err error) {
		log.Info().Str("addr", c.RemoteAddr().String()).Err(err).Msg("connection closed")
		cancel()
	})
}
