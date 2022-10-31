package kafkaconsumer

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/Shopify/sarama"
	"github.com/go-playground/validator/v10"
	"github.com/loopfz/gadgeto/zesty"
	"github.com/sirupsen/logrus"

	"github.com/ovh/utask"
	"github.com/ovh/utask/models/tasktemplate"
	"github.com/ovh/utask/pkg/auth"
	"github.com/ovh/utask/pkg/taskutils"
)

const (
	TimeoutDefault = "10s"
	ConfigKey      = "kafka-consumer"
)

// KafkaConfig is the configuration needed to write a message on Kafka topic
type KafkaConfig struct {
	Brokers      []string `json:"brokers" validate:"required,gt=0"`
	KafkaVersion string   `json:"kafka_version,omitempty"`
	Group        string   `json:"group" validate:"required,gt=0"`
	SASL         struct {
		User     string `json:"user,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"sasl,omitempty"`
	WithTLS      bool     `json:"with_tls"`
	Timeout      string   `json:"timeout"`
	Topics       []string `json:"topics" validate:"required,gt=0"`
	OldestOffset bool     `json:"oldest_offset"`

	TaskTemplate      string                 `json:"task_template" validate:"required,gt=0"`
	Input             map[string]interface{} `json:"input"`
	RequesterUsername string                 `json:"requester_username"`
	RequesterGroups   []string               `json:"requester_groups"`
	ResolverUsername  string                 `json:"resolver_username"`
	WatcherUsernames  []string               `json:"watcher_usernames"`
	WatcherGroups     []string               `json:"watcher_groups"`
}

func StartNewTaskConsumer(ctx context.Context, cfg KafkaConfig) (*Consumer, error) {
	err := validator.New().Struct(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.Timeout == "" {
		cfg.Timeout = TimeoutDefault
	}

	td, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timeout: %s", err)
	}

	// Kafka config
	config := sarama.NewConfig()
	config.Net.TLS.Enable = cfg.WithTLS
	config.Net.DialTimeout = td
	config.ClientID = "uTask-" + clearString(utask.AppName())

	if cfg.KafkaVersion != "" {
		version, err := sarama.ParseKafkaVersion(cfg.KafkaVersion)
		if err != nil {
			return nil, fmt.Errorf("failed parsing Kafka version: %v", err)
		}

		config.Version = version
	}

	// SASL authentication
	if cfg.SASL.User != "" || cfg.SASL.Password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = cfg.SASL.User
		config.Net.SASL.Password = cfg.SASL.Password
	}

	if cfg.OldestOffset {
		config.Consumer.Offsets.Initial = sarama.OffsetOldest
	}

	dbp, err := zesty.NewDBProvider(utask.DBName)
	if err != nil {
		return nil, err
	}

	tt, err := tasktemplate.LoadFromName(dbp, cfg.TaskTemplate)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		cfg:       cfg,
		saramaCfg: config,
		tt:        *tt,
	}, nil
}

func (c *Consumer) SetDefaultConsumer(ctx context.Context) error {
	client, err := sarama.NewConsumerGroup(c.cfg.Brokers, c.cfg.Group, c.saramaCfg)
	if err != nil {
		return fmt.Errorf("failed creating consumer group client: %v", err)
	}

	c.client = client

	return nil
}

func (c *Consumer) SetCustomConsumer(client sarama.ConsumerGroup) {
	c.client = client
}

func (c *Consumer) StartConsumer(ctx context.Context) {
	go func() {
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := c.client.Consume(ctx, c.cfg.Topics, c); err != nil {
				logrus.WithError(err).Warn("kafkaconsumer: fail to consume")
				continue
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	logrus.Debugf("kafkaconsumer: starting consumption")

	go func(ctx context.Context) {
		<-ctx.Done()
		if err := c.client.Close(); err != nil {
			logrus.WithError(err).Warn("kafkaconsumer: fail to close client")
		}
	}(ctx)
}

// Consumer represents a Sarama consumer group consumer
type Consumer struct {
	cfg       KafkaConfig
	tt        tasktemplate.TaskTemplate
	saramaCfg *sarama.Config
	client    sarama.ConsumerGroup
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			ctx := auth.WithIdentity(context.Background(), consumer.cfg.RequesterUsername)
			ctx = auth.WithGroups(ctx, consumer.cfg.RequesterGroups)

			dbp, err := zesty.NewDBProvider(utask.DBName)
			if err != nil {
				return err
			}

			_, err = taskutils.CreateTask(ctx, dbp,
				&consumer.tt, consumer.cfg.WatcherUsernames, consumer.cfg.WatcherGroups,
				[]string{}, []string{},
				consumer.cfg.Input, nil,
				"created from KafkaConsumer", nil, nil)
			if err != nil {
				log.Print(err)
				return err
			}

			session.MarkMessage(message, "")
			session.Commit()

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *Consumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

var (
	nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9-]+`)
)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}
