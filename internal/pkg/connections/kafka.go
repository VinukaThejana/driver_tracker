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

func getDialer(e *env.Env) *kafka.Dialer {
	return &kafka.Dialer{
		SASLMechanism: getMechanism(e),
		TLS:           &tls.Config{},
	}
}

// InitKafkaWriters is a function that is used to initialize kafka writers
func (c *C) InitKafkaWriters(e *env.Env) {
	c.K = &KafkaWriters{
		B: writer(e, e.Topic),
	}
}

// GetKafkaConnection is a function that is used to initialize the kafka connection
func (c *C) GetKafkaConnection(e *env.Env) (*kafka.Conn, error) {
	conn, err := getDialer(e).Dial("tcp", e.KafkaBroker)
	if err != nil {
		return nil, err
	}
	return conn, err
}

// GetKafkaLeaderConnection is a function that is used to get the kafka leader connection
func (c *C) GetKafkaLeaderConnection(ctx context.Context, e *env.Env, topic string, partition int) (*kafka.Conn, error) {
	conn, err := getDialer(e).DialLeader(ctx, "tcp", e.KafkaBroker, topic, partition)
	return conn, err
}

// GetLastOffset is a function that is used to get the last offset of a given partition in a given topic
func (c *C) GetLastOffset(ctx context.Context, e *env.Env, topic string, partition int) (offset int64, err error) {
	conn, err := c.GetKafkaLeaderConnection(ctx, e, topic, partition)
	if err != nil {
		return -1, err
	}
	defer conn.Close()

	offset, err = conn.ReadLastOffset()
	return offset, err
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

// GetLastNMessages is a function that is used to get the last N messages from the kafka topic within a given kafka partition
func (c *C) GetLastNMessages(ctx context.Context, e *env.Env, from int64, topic string, partition int) ([]string, error) {
	offset, err := c.GetLastOffset(ctx, e, topic, partition)
	if err != nil {
		return []string{}, err
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{e.KafkaBroker},
		Topic:   e.Topic,
		Dialer: &kafka.Dialer{
			SASLMechanism: getMechanism(e),
			TLS:           &tls.Config{},
		},
		Partition: partition,
	})
	defer reader.Close()

	err = reader.SetOffset(from)
	if err != nil {
		return []string{}, err
	}

	messages := []string{}

	for {
		m, err := reader.ReadMessage(ctx)
		if err != nil {
			return []string{}, err
		}

		messages = append(messages, string(m.Value))

		if m.Offset == offset-1 {
			break
		}
	}

	return messages, nil
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
	reader.SetOffset(offset)

	return reader
}
