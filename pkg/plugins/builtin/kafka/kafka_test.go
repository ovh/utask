package kafka

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetBrokers(t *testing.T) {
	tests := []struct {
		name    string
		brokers string
		want    []string
	}{
		{name: "Zero broker", brokers: "", want: []string{""}},
		{name: "One broker", brokers: "localhost:9092", want: []string{"localhost:9092"}},
		{name: "Two brokers", brokers: "localhost:9092,localhost:9093", want: []string{"localhost:9092", "localhost:9093"}},
		{name: "Bad separator", brokers: "localhost:9092;localhost:9093", want: []string{"localhost:9092;localhost:9093"}},
	}

	cfg := KafkaConfig{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cfg.Brokers = test.brokers

			actual := cfg.GetBrokers()
			assert.Equal(t, test.want, actual)
		})
	}
}

func Test_validConfig(t *testing.T) {
	nbExpectedBrokers := 2
	cfg := KafkaConfig{
		Brokers:      "localhost:9092,localhost:9093",
		KafkaVersion: "1.0.0.0",
		Timeout:      TimeoutDefault,
		Message: Message{
			Topic: "utask",
			Value: "hello_world",
		},
	}

	baseConfig := json.RawMessage("")
	cfgJSON, err := json.Marshal(cfg)
	assert.NoError(t, err)
	assert.NoError(t, Plugin.ValidConfig(baseConfig, cfgJSON))
	assert.Equal(t, nbExpectedBrokers, len(cfg.GetBrokers()))

	// Wrong timeout
	saveTimeout := cfg.Timeout
	cfg.Timeout = "wrong"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "timeout parameter: invalid duration")
	cfg.Timeout = saveTimeout

	// === brokers parameter tests ===
	saveBroker := cfg.Brokers

	// brokers: Empty
	cfg.Brokers = ""
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "brokers parameter: missing or empty")

	// brokers: prefixed by a scheme
	cfg.Brokers = "http://localhost:9092"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "brokers parameter: prefixed by a scheme")

	// brokers: missing port
	cfg.Brokers = "localhost:9092,localhost"
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "brokers parameter: missing port")

	cfg.Brokers = saveBroker
	// === END - brokers parameter tests ===

	// === Message parameter tests ===
	saveMessage := cfg.Message

	// Missing topic
	cfg.Message = Message{Value: saveMessage.Value}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "Message parameter: missing topic")

	// Missing value
	cfg.Message = Message{Topic: saveMessage.Topic}
	cfgJSON, err = json.Marshal(cfg)
	assert.NoError(t, err)
	assert.Errorf(t, Plugin.ValidConfig(baseConfig, cfgJSON), "Message parameter: missing value")

	cfg.Message = saveMessage
	// === END - Message parameter tests ===
}
