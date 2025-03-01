package config

// KafkaConfig holds configuration for Kafka broker and topics
type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"`
	CommandTopic  string   `yaml:"command_topic"`
	StatusTopic   string   `yaml:"status_topic"`
	ConsumerGroup string   `yaml:"consumer_group"`
}
