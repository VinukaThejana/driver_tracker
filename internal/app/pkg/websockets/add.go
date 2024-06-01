package websockets

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/connections"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/go-chi/chi/v5"
	"github.com/lesismal/nbio/nbhttp/websocket"
	"github.com/rs/zerolog/log"
)

// Add contains the websocket to add to the connected topic
type Add struct {
	E *env.Env
	C *connections.C
}

// Method of the add method
func (add *Add) Method() string {
	return http.MethodGet
}

// Path is the path of the add websocket
func (add *Add) Path() string {
	return "/add/{topic}"
}

// Handler is the bussiness logic of the add websocket
func (add *Add) Handler(w http.ResponseWriter, r *http.Request) {
	topic := chi.URLParam(r, "topic")
	if topic == "" {
		log.Error().Msg("topic is not provided")
		return
	}

	upgrader := websocket.NewUpgrader()
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}

	upgrader.OnMessage(func(_ *websocket.Conn, _ websocket.MessageType, b []byte) {
		if !json.Valid(b) {
			log.Error().Msg("provided data is not valid json")
			return
		}
		data := strings.TrimSuffix(string(b), "}") + fmt.Sprintf(`, "timestamp": %d`, time.Now().UTC().Unix()) + "}"
		add.C.KafkaWriteToTopicWithHTTP(add.E, topic, data)
	})
	upgrader.OnOpen(func(c *websocket.Conn) {
		log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection opened")
		c.OnClose(func(c *websocket.Conn, err error) {
			if err != nil {
				log.Error().Err(err).Str("addr", c.RemoteAddr().String()).Msg("connection closed with error")
			} else {
				log.Info().Str("addr", c.RemoteAddr().String()).Msg("connection closed")
			}
		})
	})

	_, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error().Err(err).Msg("error occured while upgrading the websocket connection")
		return
	}
}
