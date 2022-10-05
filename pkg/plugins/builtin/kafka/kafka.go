package kafka

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Shopify/sarama"

	"github.com/ovh/utask/pkg/plugins/taskplugin"
)

// the Kafka plugin produces a Kafka message
var (
	Plugin = taskplugin.New("kafka", "1.0", exec,
		taskplugin.WithConfig(validConfig, KafkaConfig{}),
		taskplugin.WithResources(resourcesKafka),
	)
)

const (
	// TimeoutDefault represents the default value that will be used for the request, if not defined in configuration
	TimeoutDefault    = "10s"
	DefaultMaxRetries = 5
)

type Message struct {
	Topic string `json:"topic"`
	Key   string `json:"key,omitempty"`
	Value string `json:"value"`
}

// KafkaConfig is the configuration needed to write a message on Kafka topic
type KafkaConfig struct {
	Brokers      []string `json:"brokers"`
	KafkaVersion string   `json:"kafka_version,omitempty"`
	SASL         struct {
		User     string `json:"user,omitempty"`
		Password string `json:"password,omitempty"`
	} `json:"sasl,omitempty"`
	WithTLS bool    `json:"with_tls"`
	Timeout string  `json:"timeout"`
	Message Message `json:"message"`
}

func validConfig(config interface{}) error {
	cfg := config.(*KafkaConfig)

	if len(cfg.Brokers) < 1 {
		return errors.New("missing brokers parameter")
	}

	for _, b := range cfg.Brokers {
		if b == "" {
			return errors.New("an item of the brokers list is empty")
		}

		u, err := url.Parse("http://" + b)
		if err != nil {
			return fmt.Errorf("failed to parse broker: %s", err)
		}

		if u.Port() == "" {
			return fmt.Errorf("missing port in address: %s", b)
		}
	}

	if cfg.Timeout != "" {
		if _, err := time.ParseDuration(cfg.Timeout); err != nil {
			return fmt.Errorf("failed to parse timeout parameter: %s", err)
		}
	}

	if cfg.Message.Topic == "" {
		return errors.New("missing message.topic parameter")
	}

	if cfg.Message.Value == "" {
		return errors.New("missing message.value parameter")
	}

	return nil
}

func resourcesKafka(config interface{}) []string {
	cfg := config.(*KafkaConfig)
	resources := []string{
		"socket",
	}

	exist := make(map[string]struct{})

	for _, broker := range cfg.Brokers {
		s := strings.Split(broker, ":")
		hostname := s[0]

		if _, ok := exist[hostname]; !ok {
			resources = append(resources, "url:"+hostname)
			exist[hostname] = struct{}{}
		}
	}

	return resources
}

func getKafkaConfig(cfg *KafkaConfig) (*sarama.Config, error) {
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
	config.Version = sarama.DefaultVersion

	// SASL authentication
	if cfg.SASL.User != "" || cfg.SASL.Password != "" {
		config.Net.SASL.Enable = true
		config.Net.SASL.User = cfg.SASL.User
		config.Net.SASL.Password = cfg.SASL.Password
	}

	if cfg.KafkaVersion != "" {
		kafkaVersion, err := sarama.ParseKafkaVersion(cfg.KafkaVersion)
		if err != nil {
			return config, fmt.Errorf("error parsing Kafka version %v err: %w", kafkaVersion, err)
		}
		config.Version = kafkaVersion
	}

	// Producer config
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Producer.Retry.Max = DefaultMaxRetries

	return config, nil
}

func exec(stepName string, config interface{}, ctx interface{}) (interface{}, interface{}, error) {
	cfg := config.(*KafkaConfig)

	kafkaConfig, err := getKafkaConfig(cfg)
	if err != nil {
		return nil, nil, err
	}

	producer, err := sarama.NewSyncProducer(cfg.Brokers, kafkaConfig)
	if err != nil {
		return nil, nil, err
	}
	defer producer.Close()

	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic: cfg.Message.Topic,
		Key:   sarama.ByteEncoder(cfg.Message.Key),
		Value: sarama.ByteEncoder(cfg.Message.Value),
	})
	if err != nil {
		return nil, nil, err
	}

	return map[string]interface{}{
		"partition": partition,
		"offset":    offset,
	}, nil, nil
}
