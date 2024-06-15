package connections

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/flitlabs/spotoncars-stream-go/internal/pkg/env"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"
)

// ErrKafkaNoTopic is an error that occurs when there is no topic for given driver
var ErrKafkaNoTopic = fmt.Errorf("there is no stream associated with the given driver")

func getMechanism(e *env.Env) sasl.Mechanism {
	mechanism, _ := scram.Mechanism(scram.SHA512, e.KafkaUsername, e.KafkaPassword)
	return mechanism
}

// KafkaWriters contains kafka writers
type KafkaWriters struct {
	B *kafka.Writer
}

func writer(e *env.Env, topic string) *kafka.Writer {
	w := kafka.Writer{
		Addr:  kafka.TCP(e.KafkaBroker),
		Topic: topic,
		Transport: &kafka.Transport{
			SASL: getMechanism(e),
			TLS:  &tls.Config{},
		},
		AllowAutoTopicCreation: false,
	}
	return &w
}

// InitKafkaWriters is a function that is used to initialize kafka writers
func (c *C) InitKafkaWriters(e *env.Env) {
	c.K = &KafkaWriters{
		B: writer(e, e.Topic),
	}
}

// GetKafkaConnection is a function that is used to initialize the kafka connection
func (c *C) GetKafkaConnection(e *env.Env) (*kafka.Dialer, *kafka.Conn, error) {
	dialer := &kafka.Dialer{
		SASLMechanism: getMechanism(e),
		TLS:           &tls.Config{},
	}

	conn, err := dialer.Dial("tcp", e.KafkaBroker)
	if err != nil {
		return nil, nil, err
	}
	return dialer, conn, nil
}

// KafkaWriteToTopic is a function that is used to write to a given Kafka topic
func (c *C) KafkaWriteToTopic(e *env.Env, topic string, payload []kafka.Message) {
	w := kafka.Writer{
		Addr:  kafka.TCP(e.KafkaBroker),
		Topic: topic,
		Transport: &kafka.Transport{
			SASL: getMechanism(e),
			TLS:  &tls.Config{},
		},
		AllowAutoTopicCreation: false,
		Balancer: kafka.BalancerFunc(func(m kafka.Message, i ...int) int {
			return 9
		}),
	}
	w.WriteMessages(context.Background(), payload...)
}

// KafkaWriteToTopicWithHTTP is a function that is used to write to a given Kafka topic
func (c *C) KafkaWriteToTopicWithHTTP(e *env.Env, topic string, payload string) error {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/produce/%s/%s", e.KafkaRestURL, topic, payload), nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(e.KafkaUsername, e.KafkaPassword)

	client := &http.Client{
		Timeout: time.Second,
	}
	res, err := client.Do(req)
	if err != nil {
		return ErrKafkaNoTopic
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return ErrKafkaNoTopic
	}
	return nil
}

// KafkaReader is a function that is used to intitialize a kafka reader instance
func (c *C) KafkaReader(e *env.Env, topic string, partition int, offset int64) *kafka.Reader {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{e.KafkaBroker},
		Topic:   topic,
		Dialer: &kafka.Dialer{
			SASLMechanism: getMechanism(e),
			TLS:           &tls.Config{},
		},
		Partition: partition,
	})
	reader.SetOffset(kafka.LastOffset)

	return reader
}
