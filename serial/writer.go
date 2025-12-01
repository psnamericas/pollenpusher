package serial

import (
	"fmt"
	"time"

	"go.bug.st/serial"
)

// RealPort implements Port using a real serial port
type RealPort struct {
	port   serial.Port
	config PortConfig
	isOpen bool
}

// Open opens a serial port with the given configuration
func Open(config PortConfig) (*RealPort, error) {
	mode := &serial.Mode{
		BaudRate: config.BaudRate,
		DataBits: config.DataBits,
		StopBits: convertStopBits(config.StopBits),
		Parity:   convertParity(config.Parity),
	}

	port, err := serial.Open(config.Device, mode)
	if err != nil {
		return nil, fmt.Errorf("failed to open serial port %s: %w", config.Device, err)
	}

	// Set read/write timeouts
	if err := port.SetReadTimeout(time.Second * 5); err != nil {
		port.Close()
		return nil, fmt.Errorf("failed to set read timeout: %w", err)
	}

	return &RealPort{
		port:   port,
		config: config,
		isOpen: true,
	}, nil
}

// Write writes data to the serial port
func (p *RealPort) Write(data []byte) (int, error) {
	if !p.isOpen {
		return 0, fmt.Errorf("port is closed")
	}
	return p.port.Write(data)
}

// Close closes the serial port
func (p *RealPort) Close() error {
	if !p.isOpen {
		return nil
	}
	p.isOpen = false
	return p.port.Close()
}

// Flush waits until all output has been transmitted
func (p *RealPort) Flush() error {
	if !p.isOpen {
		return fmt.Errorf("port is closed")
	}
	return p.port.Drain()
}

// Device returns the device path
func (p *RealPort) Device() string {
	return p.config.Device
}

// IsOpen returns true if the port is currently open
func (p *RealPort) IsOpen() bool {
	return p.isOpen
}

// ListPorts returns a list of available serial ports
func ListPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, fmt.Errorf("failed to list serial ports: %w", err)
	}
	return ports, nil
}

func convertStopBits(bits int) serial.StopBits {
	switch bits {
	case 1:
		return serial.OneStopBit
	case 2:
		return serial.TwoStopBits
	default:
		return serial.OneStopBit
	}
}

func convertParity(parity string) serial.Parity {
	switch parity {
	case "odd":
		return serial.OddParity
	case "even":
		return serial.EvenParity
	case "mark":
		return serial.MarkParity
	case "space":
		return serial.SpaceParity
	default:
		return serial.NoParity
	}
}
