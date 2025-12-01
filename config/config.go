package config

import (
	"encoding/json"
	"os"
	"time"
)

// Config is the root configuration structure
type Config struct {
	App        AppConfig        `json:"app"`
	Ports      []PortConfig     `json:"ports"`
	Timing     TimingConfig     `json:"timing"`
	Logging    LoggingConfig    `json:"logging"`
	Monitoring MonitoringConfig `json:"monitoring"`
	Slack      SlackConfig      `json:"slack"`
	Recovery   RecoveryConfig   `json:"recovery"`
}

// AppConfig contains application metadata
type AppConfig struct {
	Name       string `json:"name"`
	InstanceID string `json:"instance_id"`
}

// PortConfig defines configuration for a single serial port
type PortConfig struct {
	Device         string           `json:"device"`
	BaudRate       int              `json:"baud_rate"`
	DataBits       int              `json:"data_bits"`
	StopBits       int              `json:"stop_bits"`
	Parity         string           `json:"parity"`
	Format         string           `json:"format"`
	Mode           string           `json:"mode"`
	SampleFile     string           `json:"sample_file,omitempty"`
	Loop           bool             `json:"loop,omitempty"`
	CallsPerMinute float64          `json:"calls_per_minute"`
	Enabled        bool             `json:"enabled"`
	Description    string           `json:"description,omitempty"`
	Synthetic      *SyntheticConfig `json:"synthetic,omitempty"`
}

// SyntheticConfig contains settings for synthetic data generation
type SyntheticConfig struct {
	SystemID           string `json:"system_id"`
	AgentCount         int    `json:"agent_count"`
	MinDurationSec     int    `json:"min_duration_sec"`
	MaxDurationSec     int    `json:"max_duration_sec"`
	IncludeAgentEvents bool   `json:"include_agent_events"`
}

// TimingConfig controls timing behavior
type TimingConfig struct {
	JitterPercent   float64 `json:"jitter_percent"`
	StartupDelaySec int     `json:"startup_delay_sec"`
}

// LoggingConfig defines logging settings
type LoggingConfig struct {
	Level      string `json:"level"`
	BasePath   string `json:"base_path"`
	Filename   string `json:"filename"`
	MaxSizeMB  int    `json:"max_size_mb"`
	MaxBackups int    `json:"max_backups"`
	Compress   bool   `json:"compress"`
}

// MonitoringConfig defines HTTP monitoring settings
type MonitoringConfig struct {
	Port             int `json:"port"`
	StatsIntervalSec int `json:"stats_interval_sec"`
}

// SlackConfig defines Slack notification settings
type SlackConfig struct {
	WebhookURL     string `json:"webhook_url"`
	NotifyStartup  bool   `json:"notify_startup"`
	NotifyShutdown bool   `json:"notify_shutdown"`
	NotifyErrors   bool   `json:"notify_errors"`
}

// RecoveryConfig defines reconnection behavior
type RecoveryConfig struct {
	ReconnectDelaySec    int  `json:"reconnect_delay_sec"`
	MaxReconnectDelaySec int  `json:"max_reconnect_delay_sec"`
	ExponentialBackoff   bool `json:"exponential_backoff"`
}

// Load reads and parses a configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// Apply defaults
	cfg.applyDefaults()

	return &cfg, nil
}

// applyDefaults sets default values for unspecified fields
func (c *Config) applyDefaults() {
	// App defaults
	if c.App.Name == "" {
		c.App.Name = "CDRGenerator"
	}
	if c.App.InstanceID == "" {
		hostname, _ := os.Hostname()
		c.App.InstanceID = hostname
	}

	// Port defaults
	for i := range c.Ports {
		if c.Ports[i].BaudRate == 0 {
			c.Ports[i].BaudRate = 9600
		}
		if c.Ports[i].DataBits == 0 {
			c.Ports[i].DataBits = 8
		}
		if c.Ports[i].StopBits == 0 {
			c.Ports[i].StopBits = 1
		}
		if c.Ports[i].Parity == "" {
			c.Ports[i].Parity = "none"
		}
		if c.Ports[i].CallsPerMinute == 0 {
			c.Ports[i].CallsPerMinute = 1.0
		}
	}

	// Timing defaults
	if c.Timing.JitterPercent == 0 {
		c.Timing.JitterPercent = 10.0
	}

	// Logging defaults
	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Filename == "" {
		c.Logging.Filename = "cdrgenerator.log"
	}
	if c.Logging.MaxSizeMB == 0 {
		c.Logging.MaxSizeMB = 50
	}
	if c.Logging.MaxBackups == 0 {
		c.Logging.MaxBackups = 5
	}

	// Monitoring defaults
	if c.Monitoring.Port == 0 {
		c.Monitoring.Port = 8080
	}
	if c.Monitoring.StatsIntervalSec == 0 {
		c.Monitoring.StatsIntervalSec = 60
	}

	// Recovery defaults
	if c.Recovery.ReconnectDelaySec == 0 {
		c.Recovery.ReconnectDelaySec = 5
	}
	if c.Recovery.MaxReconnectDelaySec == 0 {
		c.Recovery.MaxReconnectDelaySec = 300
	}
}

// GetReconnectDelay returns the initial reconnect delay as a duration
func (c *RecoveryConfig) GetReconnectDelay() time.Duration {
	return time.Duration(c.ReconnectDelaySec) * time.Second
}

// GetMaxReconnectDelay returns the maximum reconnect delay as a duration
func (c *RecoveryConfig) GetMaxReconnectDelay() time.Duration {
	return time.Duration(c.MaxReconnectDelaySec) * time.Second
}

// GetStartupDelay returns the startup delay as a duration
func (c *TimingConfig) GetStartupDelay() time.Duration {
	return time.Duration(c.StartupDelaySec) * time.Second
}
