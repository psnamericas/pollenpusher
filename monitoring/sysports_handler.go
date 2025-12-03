package monitoring

import (
	"bufio"
	"encoding/json"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// SysPortInfo contains system-level serial port information
type SysPortInfo struct {
	Device   string `json:"device"`
	UART     string `json:"uart"`
	Port     string `json:"port"`
	IRQ      int    `json:"irq"`
	TX       int64  `json:"tx"`
	RX       int64  `json:"rx"`
	Signals  string `json:"signals"`
	Active   bool   `json:"active"`
	COMPort  string `json:"com_port"`
}

// SysPortsHandler handles requests for system serial port info
type SysPortsHandler struct{}

// NewSysPortsHandler creates a new system ports handler
func NewSysPortsHandler() *SysPortsHandler {
	return &SysPortsHandler{}
}

// ServeHTTP handles system port info requests
func (h *SysPortsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ports, err := h.getSystemPorts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"ports": ports,
	})
}

func (h *SysPortsHandler) getSystemPorts() ([]SysPortInfo, error) {
	file, err := os.Open("/proc/tty/driver/serial")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var ports []SysPortInfo
	scanner := bufio.NewScanner(file)

	// Regex to parse serial port lines
	// Example: "4: uart:16550A port:000002F0 irq:7 tx:1195 rx:1170 CTS|DSR|CD"
	re := regexp.MustCompile(`^\s*(\d+):\s+uart:(\S+)\s+port:([0-9A-Fa-f]+)\s+irq:(\d+)\s+tx:(\d+)\s+rx:(\d+)(.*)$`)

	for scanner.Scan() {
		line := scanner.Text()
		matches := re.FindStringSubmatch(line)

		if len(matches) >= 7 {
			portNum, _ := strconv.Atoi(matches[1])
			irq, _ := strconv.Atoi(matches[4])
			tx, _ := strconv.ParseInt(matches[5], 10, 64)
			rx, _ := strconv.ParseInt(matches[6], 10, 64)

			signals := strings.TrimSpace(matches[7])

			// Determine COM port number (COM1 = ttyS0, etc.)
			comPort := ""
			if matches[2] != "unknown" {
				comPort = "COM" + strconv.Itoa(portNum+1)
			}

			// Port is considered active (cable connected) if:
			// 1. It has CTS/DSR/CD signals (remote device present), OR
			// 2. Both TX and RX are non-zero (bidirectional communication)
			hasCTS := strings.Contains(signals, "CTS")
			hasDSR := strings.Contains(signals, "DSR")
			hasCD := strings.Contains(signals, "CD")
			hasRemoteSignals := hasCTS || hasDSR || hasCD
			hasBidirectional := tx > 0 && rx > 0
			active := matches[2] != "unknown" && (hasRemoteSignals || hasBidirectional)

			port := SysPortInfo{
				Device:  "/dev/ttyS" + strconv.Itoa(portNum),
				UART:    matches[2],
				Port:    "0x" + strings.ToUpper(matches[3]),
				IRQ:     irq,
				TX:      tx,
				RX:      rx,
				Signals: signals,
				Active:  active,
				COMPort: comPort,
			}

			// Only include hardware ports (skip unknown)
			if matches[2] != "unknown" {
				ports = append(ports, port)
			}
		}
	}

	return ports, scanner.Err()
}
