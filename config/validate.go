package config

import (
	"fmt"
	"os"
	"strings"
)

// ValidationError contains details about configuration validation failures
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Validate checks the configuration for errors
func Validate(cfg *Config, availableFormats []string) error {
	var errors ValidationErrors

	// Validate ports
	if len(cfg.Ports) == 0 {
		errors = append(errors, ValidationError{
			Field:   "ports",
			Message: "at least one port must be configured",
		})
	}

	devicesSeen := make(map[string]bool)
	for i, port := range cfg.Ports {
		portErrors := validatePort(port, i, availableFormats, devicesSeen)
		errors = append(errors, portErrors...)
	}

	// Validate timing
	if cfg.Timing.JitterPercent < 0 || cfg.Timing.JitterPercent > 100 {
		errors = append(errors, ValidationError{
			Field:   "timing.jitter_percent",
			Message: "must be between 0 and 100",
		})
	}

	// Validate logging
	if cfg.Logging.BasePath != "" {
		if info, err := os.Stat(cfg.Logging.BasePath); err != nil || !info.IsDir() {
			errors = append(errors, ValidationError{
				Field:   "logging.base_path",
				Message: fmt.Sprintf("directory does not exist: %s", cfg.Logging.BasePath),
			})
		}
	}

	// Validate monitoring
	if cfg.Monitoring.Port < 1 || cfg.Monitoring.Port > 65535 {
		errors = append(errors, ValidationError{
			Field:   "monitoring.port",
			Message: "must be between 1 and 65535",
		})
	}

	// Validate recovery
	if cfg.Recovery.ReconnectDelaySec < 1 {
		errors = append(errors, ValidationError{
			Field:   "recovery.reconnect_delay_sec",
			Message: "must be at least 1 second",
		})
	}

	if len(errors) > 0 {
		return errors
	}
	return nil
}

func validatePort(port PortConfig, index int, availableFormats []string, devicesSeen map[string]bool) ValidationErrors {
	var errors ValidationErrors
	prefix := fmt.Sprintf("ports[%d]", index)

	// Check device
	if port.Device == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".device",
			Message: "device path is required",
		})
	} else if devicesSeen[port.Device] {
		errors = append(errors, ValidationError{
			Field:   prefix + ".device",
			Message: fmt.Sprintf("duplicate device: %s", port.Device),
		})
	} else {
		devicesSeen[port.Device] = true
	}

	// Check baud rate
	validBaudRates := []int{300, 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200}
	if !contains(validBaudRates, port.BaudRate) {
		errors = append(errors, ValidationError{
			Field:   prefix + ".baud_rate",
			Message: fmt.Sprintf("invalid baud rate: %d", port.BaudRate),
		})
	}

	// Check format
	if port.Format == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".format",
			Message: "format is required",
		})
	} else if !containsString(availableFormats, strings.ToLower(port.Format)) {
		errors = append(errors, ValidationError{
			Field:   prefix + ".format",
			Message: fmt.Sprintf("unknown format: %s (available: %s)", port.Format, strings.Join(availableFormats, ", ")),
		})
	}

	// Check mode
	validModes := []string{"replay", "synthetic"}
	if !containsString(validModes, strings.ToLower(port.Mode)) {
		errors = append(errors, ValidationError{
			Field:   prefix + ".mode",
			Message: fmt.Sprintf("invalid mode: %s (must be 'replay' or 'synthetic')", port.Mode),
		})
	}

	// Mode-specific validation
	if strings.ToLower(port.Mode) == "replay" {
		if port.SampleFile == "" {
			errors = append(errors, ValidationError{
				Field:   prefix + ".sample_file",
				Message: "sample_file is required for replay mode",
			})
		} else if _, err := os.Stat(port.SampleFile); os.IsNotExist(err) {
			errors = append(errors, ValidationError{
				Field:   prefix + ".sample_file",
				Message: fmt.Sprintf("file does not exist: %s", port.SampleFile),
			})
		}
	}

	if strings.ToLower(port.Mode) == "synthetic" {
		if port.Synthetic == nil {
			errors = append(errors, ValidationError{
				Field:   prefix + ".synthetic",
				Message: "synthetic configuration is required for synthetic mode",
			})
		} else {
			synthErrors := validateSynthetic(port.Synthetic, prefix)
			errors = append(errors, synthErrors...)
		}
	}

	// Check calls per minute
	if port.CallsPerMinute <= 0 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".calls_per_minute",
			Message: "must be greater than 0",
		})
	}

	return errors
}

func validateSynthetic(synth *SyntheticConfig, prefix string) ValidationErrors {
	var errors ValidationErrors

	if synth.SystemID == "" {
		errors = append(errors, ValidationError{
			Field:   prefix + ".synthetic.system_id",
			Message: "system_id is required",
		})
	}

	if synth.AgentCount < 1 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".synthetic.agent_count",
			Message: "must be at least 1",
		})
	}

	if synth.MinDurationSec < 1 {
		errors = append(errors, ValidationError{
			Field:   prefix + ".synthetic.min_duration_sec",
			Message: "must be at least 1 second",
		})
	}

	if synth.MaxDurationSec < synth.MinDurationSec {
		errors = append(errors, ValidationError{
			Field:   prefix + ".synthetic.max_duration_sec",
			Message: "must be greater than or equal to min_duration_sec",
		})
	}

	return errors
}

func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
