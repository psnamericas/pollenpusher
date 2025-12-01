package output

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"cdrgenerator/config"
	"cdrgenerator/generator"
	"cdrgenerator/serial"
)

// ChannelState represents the current state of an output channel
type ChannelState string

const (
	StateInitializing ChannelState = "initializing"
	StateRunning      ChannelState = "running"
	StatePaused       ChannelState = "paused"
	StateReconnecting ChannelState = "reconnecting"
	StateStopped      ChannelState = "stopped"
	StateError        ChannelState = "error"
)

// Channel manages output to a single serial port
type Channel struct {
	config     *config.PortConfig
	recovery   *config.RecoveryConfig
	generator  *generator.Generator
	port       serial.Port
	portStats  *serial.PortWithStats
	logger     *slog.Logger

	state      ChannelState
	stateMutex sync.RWMutex

	// Statistics
	stats      ChannelStats
	statsMutex sync.RWMutex

	// Control
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// ChannelStats contains statistics for an output channel
type ChannelStats struct {
	RecordsSent    int64
	BytesSent      int64
	Errors         int64
	LastRecordTime time.Time
	StartTime      time.Time
	LastError      string
}

// NewChannel creates a new output channel
func NewChannel(
	portCfg *config.PortConfig,
	recoveryCfg *config.RecoveryConfig,
	gen *generator.Generator,
	logger *slog.Logger,
) *Channel {
	return &Channel{
		config:    portCfg,
		recovery:  recoveryCfg,
		generator: gen,
		logger:    logger.With("device", portCfg.Device, "format", portCfg.Format),
		state:     StateInitializing,
		stopCh:    make(chan struct{}),
		stats: ChannelStats{
			StartTime: time.Now(),
		},
	}
}

// Start begins the output channel
func (c *Channel) Start(ctx context.Context) error {
	c.setState(StateInitializing)

	// Open the serial port
	if err := c.openPort(); err != nil {
		c.setState(StateError)
		return fmt.Errorf("failed to open port: %w", err)
	}

	c.setState(StateRunning)
	c.logger.Info("Output channel started",
		"mode", c.generator.Mode(),
		"calls_per_minute", c.config.CallsPerMinute,
	)

	// Start the output loop
	c.wg.Add(1)
	go c.outputLoop(ctx)

	return nil
}

// Stop gracefully stops the output channel
func (c *Channel) Stop() {
	c.logger.Info("Stopping output channel")
	close(c.stopCh)
	c.wg.Wait()

	if c.port != nil {
		c.port.Close()
	}

	c.setState(StateStopped)
	c.logger.Info("Output channel stopped",
		"records_sent", c.stats.RecordsSent,
		"bytes_sent", c.stats.BytesSent,
	)
}

// State returns the current channel state
func (c *Channel) State() ChannelState {
	c.stateMutex.RLock()
	defer c.stateMutex.RUnlock()
	return c.state
}

// Stats returns a copy of the current statistics
func (c *Channel) Stats() ChannelStats {
	c.statsMutex.RLock()
	defer c.statsMutex.RUnlock()
	return c.stats
}

func (c *Channel) setState(state ChannelState) {
	c.stateMutex.Lock()
	defer c.stateMutex.Unlock()
	c.state = state
}

func (c *Channel) openPort() error {
	portCfg := serial.PortConfig{
		Device:   c.config.Device,
		BaudRate: c.config.BaudRate,
		DataBits: c.config.DataBits,
		StopBits: c.config.StopBits,
		Parity:   c.config.Parity,
	}

	port, err := serial.Open(portCfg)
	if err != nil {
		return err
	}

	c.port = port
	c.portStats = serial.NewPortWithStats(port)
	return nil
}

func (c *Channel) outputLoop(ctx context.Context) {
	defer c.wg.Done()

	ticker := generator.NewTicker(c.generator.RateLimiter())
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-c.stopCh:
			return
		case <-ticker.C:
			if err := c.sendNextRecord(ctx); err != nil {
				c.handleError(err)
			}
		}
	}
}

func (c *Channel) sendNextRecord(ctx context.Context) error {
	// Get the next record
	record, err := c.generator.NextRecord(ctx)
	if err != nil {
		return fmt.Errorf("failed to get next record: %w", err)
	}

	// Write to port
	data := record.Output()
	n, err := c.portStats.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write to port: %w", err)
	}

	// Flush to ensure data is sent
	if err := c.port.Flush(); err != nil {
		c.logger.Warn("Failed to flush port", "error", err)
	}

	// Update statistics
	c.statsMutex.Lock()
	c.stats.RecordsSent++
	c.stats.BytesSent += int64(n)
	c.stats.LastRecordTime = time.Now()
	c.statsMutex.Unlock()

	c.portStats.RecordSent()

	c.logger.Debug("Sent record",
		"record_id", record.ID,
		"bytes", n,
	)

	return nil
}

func (c *Channel) handleError(err error) {
	c.statsMutex.Lock()
	c.stats.Errors++
	c.stats.LastError = err.Error()
	c.statsMutex.Unlock()

	c.logger.Error("Output error", "error", err)

	// Check if we need to reconnect
	if !c.port.IsOpen() {
		c.reconnect()
	}
}

func (c *Channel) reconnect() {
	c.setState(StateReconnecting)

	delay := c.recovery.GetReconnectDelay()
	maxDelay := c.recovery.GetMaxReconnectDelay()
	attempt := 0

	for {
		select {
		case <-c.stopCh:
			return
		default:
		}

		attempt++
		c.logger.Info("Attempting to reconnect", "attempt", attempt, "delay", delay)

		time.Sleep(delay)

		if err := c.openPort(); err != nil {
			c.logger.Warn("Reconnection failed", "error", err)

			// Exponential backoff
			if c.recovery.ExponentialBackoff {
				delay = delay * 2
				if delay > maxDelay {
					delay = maxDelay
				}
			}
			continue
		}

		c.logger.Info("Reconnected successfully", "attempt", attempt)
		c.setState(StateRunning)
		return
	}
}

// Device returns the device path
func (c *Channel) Device() string {
	return c.config.Device
}

// Format returns the format name
func (c *Channel) Format() string {
	return c.config.Format
}

// Mode returns the mode
func (c *Channel) Mode() string {
	return c.config.Mode
}
