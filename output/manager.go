package output

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cdrgenerator/config"
	"cdrgenerator/generator"
)

// Manager manages all output channels
type Manager struct {
	config   *config.Config
	channels []*Channel
	logger   *slog.Logger
	mu       sync.RWMutex
}

// NewManager creates a new output manager
func NewManager(cfg *config.Config, logger *slog.Logger) *Manager {
	return &Manager{
		config:   cfg,
		channels: make([]*Channel, 0),
		logger:   logger,
	}
}

// Start initializes and starts all enabled output channels
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, portCfg := range m.config.Ports {
		if !portCfg.Enabled {
			m.logger.Info("Skipping disabled port", "device", portCfg.Device)
			continue
		}

		portCfgCopy := portCfg // Create a copy for the closure

		// Create generator for this port
		gen, err := generator.New(&portCfgCopy, m.config.Timing.JitterPercent)
		if err != nil {
			return fmt.Errorf("failed to create generator for %s: %w", portCfg.Device, err)
		}

		// Create output channel
		channel := NewChannel(&portCfgCopy, &m.config.Recovery, gen, m.logger)

		// Start the channel
		if err := channel.Start(ctx); err != nil {
			m.logger.Error("Failed to start channel",
				"device", portCfg.Device,
				"error", err,
			)
			// Continue with other channels
			continue
		}

		m.channels = append(m.channels, channel)
		m.logger.Info("Started output channel",
			"device", portCfg.Device,
			"format", portCfg.Format,
			"mode", portCfg.Mode,
		)
	}

	if len(m.channels) == 0 {
		return fmt.Errorf("no output channels started")
	}

	m.logger.Info("Output manager started", "channels", len(m.channels))
	return nil
}

// Stop gracefully stops all output channels
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logger.Info("Stopping output manager", "channels", len(m.channels))

	var wg sync.WaitGroup
	for _, channel := range m.channels {
		wg.Add(1)
		go func(ch *Channel) {
			defer wg.Done()
			ch.Stop()
		}(channel)
	}
	wg.Wait()

	m.logger.Info("Output manager stopped")
}

// GetStats returns statistics for all channels
func (m *Manager) GetStats() map[string]ChannelStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]ChannelStats)
	for _, channel := range m.channels {
		stats[channel.Device()] = channel.Stats()
	}
	return stats
}

// GetChannelStates returns the state of all channels
func (m *Manager) GetChannelStates() map[string]ChannelInfo {
	m.mu.RLock()
	defer m.mu.RUnlock()

	states := make(map[string]ChannelInfo)
	for _, channel := range m.channels {
		stats := channel.Stats()
		states[channel.Device()] = ChannelInfo{
			Device:         channel.Device(),
			Format:         channel.Format(),
			Mode:           channel.Mode(),
			State:          string(channel.State()),
			RecordsSent:    stats.RecordsSent,
			BytesSent:      stats.BytesSent,
			Errors:         stats.Errors,
			LastRecordTime: stats.LastRecordTime,
			LastError:      stats.LastError,
		}
	}
	return states
}

// ChannelInfo contains information about a channel for external consumers
type ChannelInfo struct {
	Device         string    `json:"device"`
	Format         string    `json:"format"`
	Mode           string    `json:"mode"`
	State          string    `json:"state"`
	RecordsSent    int64     `json:"records_sent"`
	BytesSent      int64     `json:"bytes_sent"`
	Errors         int64     `json:"errors"`
	LastRecordTime time.Time `json:"last_record_time"`
	LastError      string    `json:"last_error,omitempty"`
}

// ChannelCount returns the number of active channels
func (m *Manager) ChannelCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.channels)
}
