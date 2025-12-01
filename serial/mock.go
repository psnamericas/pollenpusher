package serial

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"
)

// MockPort implements Port for testing purposes
type MockPort struct {
	mu       sync.Mutex
	buffer   bytes.Buffer
	device   string
	isOpen   bool
	writes   [][]byte
	writeErr error // If set, Write will return this error
}

// NewMockPort creates a new mock port
func NewMockPort(device string) *MockPort {
	return &MockPort{
		device: device,
		isOpen: true,
		writes: make([][]byte, 0),
	}
}

// Write writes data to the mock port buffer
func (p *MockPort) Write(data []byte) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isOpen {
		return 0, fmt.Errorf("port is closed")
	}

	if p.writeErr != nil {
		return 0, p.writeErr
	}

	// Store a copy of the data
	dataCopy := make([]byte, len(data))
	copy(dataCopy, data)
	p.writes = append(p.writes, dataCopy)

	return p.buffer.Write(data)
}

// Close closes the mock port
func (p *MockPort) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isOpen = false
	return nil
}

// Flush is a no-op for the mock port
func (p *MockPort) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.isOpen {
		return fmt.Errorf("port is closed")
	}
	return nil
}

// Device returns the mock device path
func (p *MockPort) Device() string {
	return p.device
}

// IsOpen returns true if the mock port is open
func (p *MockPort) IsOpen() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.isOpen
}

// GetWrittenData returns all data written to the mock port
func (p *MockPort) GetWrittenData() []byte {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buffer.Bytes()
}

// GetWrites returns all individual write operations
func (p *MockPort) GetWrites() [][]byte {
	p.mu.Lock()
	defer p.mu.Unlock()

	result := make([][]byte, len(p.writes))
	for i, w := range p.writes {
		result[i] = make([]byte, len(w))
		copy(result[i], w)
	}
	return result
}

// Reset clears all written data
func (p *MockPort) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer.Reset()
	p.writes = make([][]byte, 0)
}

// SetWriteError sets an error to be returned on subsequent writes
func (p *MockPort) SetWriteError(err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writeErr = err
}

// ClearWriteError clears any write error
func (p *MockPort) ClearWriteError() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.writeErr = nil
}

// Reopen reopens a closed mock port
func (p *MockPort) Reopen() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.isOpen = true
}

// FilePort implements Port using a file (useful for testing with PTY or file output)
type FilePort struct {
	writer io.WriteCloser
	device string
	isOpen bool
}

// NewFilePort creates a new file-based port
func NewFilePort(device string, writer io.WriteCloser) *FilePort {
	return &FilePort{
		writer: writer,
		device: device,
		isOpen: true,
	}
}

// Write writes data to the file
func (p *FilePort) Write(data []byte) (int, error) {
	if !p.isOpen {
		return 0, fmt.Errorf("port is closed")
	}
	return p.writer.Write(data)
}

// Close closes the file
func (p *FilePort) Close() error {
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	return p.writer.Close()
}

// Flush is a no-op for file ports (could implement sync if needed)
func (p *FilePort) Flush() error {
	if !p.isOpen {
		return fmt.Errorf("port is closed")
	}
	// If the writer supports sync, we could call it here
	if syncer, ok := p.writer.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}

// Device returns the device/file path
func (p *FilePort) Device() string {
	return p.device
}

// IsOpen returns true if the file port is open
func (p *FilePort) IsOpen() bool {
	return p.isOpen
}

// StdoutPort implements Port writing to stdout (useful for debugging)
type StdoutPort struct {
	device string
	isOpen bool
}

// NewStdoutPort creates a new stdout port
func NewStdoutPort(device string) *StdoutPort {
	return &StdoutPort{
		device: device,
		isOpen: true,
	}
}

// Write writes data to stdout
func (p *StdoutPort) Write(data []byte) (int, error) {
	if !p.isOpen {
		return 0, fmt.Errorf("port is closed")
	}
	fmt.Printf("[%s][%s] %s", p.device, time.Now().Format("15:04:05.000"), string(data))
	return len(data), nil
}

// Close closes the stdout port
func (p *StdoutPort) Close() error {
	p.isOpen = false
	return nil
}

// Flush is a no-op for stdout
func (p *StdoutPort) Flush() error {
	return nil
}

// Device returns the device name
func (p *StdoutPort) Device() string {
	return p.device
}

// IsOpen returns true if the port is open
func (p *StdoutPort) IsOpen() bool {
	return p.isOpen
}
