package websockets

import (
	"net/http"
	"time"

	"github.com/bytedance/sonic"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
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
	topic := chi.URLParam(r, "topic")
	if topic == "" {
		http.Error(w, "provide a valid booking id", http.StatusBadRequest)
		return
	}

	isDriver := true

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	if isDriver {
		upgrader.OnMessage(func(_ *websocket.Conn, _ websocket.MessageType, b []byte) {
			var (
				data    map[string]interface{}
				payload string
				err     error
			)

			if err = sonic.UnmarshalString(string(b), &data); err != nil {
				log.Error().Err(err).Msg("provide valid JSON data")
				return
			}
			data["timestamp"] = time.Now().UTC().Unix()
			if payload, err = sonic.MarshalString(data); err != nil {
				log.Error().Err(err).Msg("failed to marshal data")
				return
			}

			if err = view.C.KafkaWriteToTopicWithHTTP(view.E, topic, payload); err != nil {
				log.Error().Err(err).Msg("failed to write data to kafka")
				return
			}
		})
	}

	upgrader.OnOpen(func(c *websocket.Conn) {
		log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection opened")
		done := make(chan struct{})

		go func() {
			reader := view.C.KafkaReader(view.E, topic, kafka.LastOffset)
			defer reader.Close()
			defer close(done)

			for {
				select {
				case <-done:
					return
				default:
					message, _ := reader.ReadMessage(r.Context())
					if len(message.Value) == 0 {
						continue
					}
					if err := c.WriteMessage(websocket.TextMessage, message.Value); err != nil {
						log.Error().Err(err).Msg("error sending data to the websocket client")
						return
					}
				}
			}
		}()

		c.OnClose(func(c *websocket.Conn, err error) {
			if err != nil {
				log.Error().Err(err).Str("addr", c.RemoteAddr().String()).Msg("connection closed with error")
			} else {
				log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection closed")
			}
			done <- struct{}{}
		})
	})

	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("error occured while upgrading the websocket connection")
		return
	}
}
