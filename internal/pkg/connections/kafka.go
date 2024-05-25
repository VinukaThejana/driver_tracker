package connections

import (
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/VinukaThejana/go-utils/logger"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
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

// KafkaReader is a function that is used to intitialize a kafka reader instance
func (c *C) KafkaReader(e *env.Env, topic string, offset int64) *kafka.Reader {
	mechanism, _ := scram.Mechanism(scram.SHA512, e.KafkaUsername, e.KafkaPassword)
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{e.KafkaBroker},
		Topic:       topic,
		StartOffset: offset,
		Dialer: &kafka.Dialer{
			SASLMechanism: mechanism,
			TLS:           &tls.Config{},
		},
	})

	return reader
}
