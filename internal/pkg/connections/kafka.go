package connections

import (
	"context"
	"crypto/tls"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/lib"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// KafkaWriteToTopic is a function that is used to write to a given Kafka topic
func (c *C) KafkaWriteToTopic(e *env.Env, topic string, payload string) {
	mechanism, err := scram.Mechanism(scram.SHA512, e.KafkaUsername, e.KafkaPassword)
	lib.LogFatal(err)

	w := kafka.Writer{
		Addr:  kafka.TCP(e.KafkaBroker),
		Topic: topic,
		Transport: &kafka.Transport{
			SASL: mechanism,
			TLS:  &tls.Config{},
		},
	}
	defer w.Close()

	w.WriteMessages(context.Background(), kafka.Message{
		Value: []byte(payload),
	})
}
