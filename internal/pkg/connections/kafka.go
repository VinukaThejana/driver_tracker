package connections

import (
	"fmt"
	"net/http"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
)

// KafkaWriteToTopic is a function that is used to write to a given Kafka topic
func (c *C) KafkaWriteToTopic(e *env.Env, topic string, payload string) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/produce/%s/%s", e.KafkaRestURL, topic, payload), nil)
	if err != nil {
		logger.ErrorWithMsg(err, "failed to create the request")
		return
	}
	req.SetBasicAuth(e.KafkaUsername, e.KafkaPassword)

	client := &http.Client{}
	_, err = client.Do(req)
	if err != nil {
		logger.ErrorWithMsg(err, "error sending the reqeust")
		return
	}
}
