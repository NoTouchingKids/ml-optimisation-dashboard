package config

// LogStreamingConfig holds configuration for the log streaming service
type LogStreamingConfig struct {
	Host          string `yaml:"host"`
	Port          int    `yaml:"port"`
	BufferSize    int    `yaml:"buffer_size"`
	FlushInterval int    `yaml:"flush_interval_ms"`
}
