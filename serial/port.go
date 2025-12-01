package serial

import (
	"io"
	"time"
)

// PortConfig contains serial port configuration settings
type PortConfig struct {
	Device   string
	BaudRate int
	DataBits int
	StopBits int
	Parity   string // "none", "odd", "even"
}

// Port defines the interface for serial port operations
type Port interface {
	io.WriteCloser

	// Flush waits until all output has been transmitted
	Flush() error

	// Device returns the device path
	Device() string

	// IsOpen returns true if the port is currently open
	IsOpen() bool
}

// Stats tracks statistics for a serial port
type Stats struct {
	BytesSent      int64
	RecordsSent    int64
	Errors         int64
	LastRecordTime time.Time
	OpenedAt       time.Time
}

// PortWithStats wraps a Port with statistics tracking
type PortWithStats struct {
	Port
	stats Stats
}

// NewPortWithStats creates a new port wrapper with statistics
func NewPortWithStats(port Port) *PortWithStats {
	return &PortWithStats{
		Port: port,
		stats: Stats{
			OpenedAt: time.Now(),
		},
	}
}

// Write writes data to the port and tracks statistics
func (p *PortWithStats) Write(data []byte) (int, error) {
	n, err := p.Port.Write(data)
	if err != nil {
		p.stats.Errors++
		return n, err
	}
	p.stats.BytesSent += int64(n)
	return n, nil
}

// RecordSent increments the records sent counter
func (p *PortWithStats) RecordSent() {
	p.stats.RecordsSent++
	p.stats.LastRecordTime = time.Now()
}

// Stats returns a copy of the current statistics
func (p *PortWithStats) Stats() Stats {
	return p.stats
}

// IncrementErrors increments the error counter
func (p *PortWithStats) IncrementErrors() {
	p.stats.Errors++
}
