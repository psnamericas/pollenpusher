package ui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// HealthResponse matches the monitoring API response
type HealthResponse struct {
	Status     string                 `json:"status"`
	InstanceID string                 `json:"instance_id"`
	Version    string                 `json:"version"`
	UptimeSec  int64                  `json:"uptime_sec"`
	Ports      map[string]PortInfo    `json:"ports"`
}

// PortInfo contains information about a port
type PortInfo struct {
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

// DashboardTab represents the dashboard UI
type DashboardTab struct {
	apiURL          string
	statusLabel     *widget.Label
	instanceLabel   *widget.Label
	versionLabel    *widget.Label
	uptimeLabel     *widget.Label
	portTable       *widget.Table
	refreshInterval time.Duration
	portData        []PortInfo
	stopRefresh     chan bool
}

// NewDashboardTab creates a new dashboard tab
func NewDashboardTab() *DashboardTab {
	return &DashboardTab{
		apiURL:          "http://localhost:8080",
		refreshInterval: 2 * time.Second,
		portData:        make([]PortInfo, 0),
		stopRefresh:     make(chan bool),
	}
}

// Build constructs the dashboard UI
func (d *DashboardTab) Build() *fyne.Container {
	// Status section
	d.statusLabel = widget.NewLabel("Status: Unknown")
	d.instanceLabel = widget.NewLabel("Instance: -")
	d.versionLabel = widget.NewLabel("Version: -")
	d.uptimeLabel = widget.NewLabel("Uptime: -")

	statusCard := widget.NewCard("Service Status", "", container.NewVBox(
		d.statusLabel,
		d.instanceLabel,
		d.versionLabel,
		d.uptimeLabel,
	))

	// Port status table
	d.portTable = widget.NewTable(
		func() (int, int) {
			return len(d.portData) + 1, 7 // +1 for header row
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, cell fyne.CanvasObject) {
			label := cell.(*widget.Label)

			// Header row
			if id.Row == 0 {
				headers := []string{"Device", "Format", "State", "Records", "Bytes", "Errors", "Last Record"}
				if id.Col < len(headers) {
					label.SetText(headers[id.Col])
					label.TextStyle = fyne.TextStyle{Bold: true}
				}
				return
			}

			// Data rows
			if id.Row-1 < len(d.portData) {
				port := d.portData[id.Row-1]
				label.TextStyle = fyne.TextStyle{}

				switch id.Col {
				case 0:
					label.SetText(port.Device)
				case 1:
					label.SetText(port.Format)
				case 2:
					label.SetText(port.State)
					// Color code the state
					switch port.State {
					case "running":
						label.Importance = widget.SuccessImportance
					case "error":
						label.Importance = widget.DangerImportance
					default:
						label.Importance = widget.MediumImportance
					}
				case 3:
					label.SetText(fmt.Sprintf("%d", port.RecordsSent))
				case 4:
					label.SetText(fmt.Sprintf("%d", port.BytesSent))
				case 5:
					label.SetText(fmt.Sprintf("%d", port.Errors))
				case 6:
					if !port.LastRecordTime.IsZero() {
						label.SetText(port.LastRecordTime.Format("15:04:05"))
					} else {
						label.SetText("-")
					}
				}
			}
		},
	)

	// Set column widths
	d.portTable.SetColumnWidth(0, 120) // Device
	d.portTable.SetColumnWidth(1, 80)  // Format
	d.portTable.SetColumnWidth(2, 100) // State
	d.portTable.SetColumnWidth(3, 80)  // Records
	d.portTable.SetColumnWidth(4, 100) // Bytes
	d.portTable.SetColumnWidth(5, 80)  // Errors
	d.portTable.SetColumnWidth(6, 100) // Last Record

	portCard := widget.NewCard("Port Status", "", container.NewScroll(d.portTable))

	// Refresh button
	refreshBtn := widget.NewButton("Refresh Now", func() {
		d.fetchHealth()
	})

	// Auto-refresh toggle
	autoRefreshCheck := widget.NewCheck("Auto-refresh (2s)", func(checked bool) {
		if checked {
			go d.startAutoRefresh()
		} else {
			d.stopRefresh <- true
		}
	})
	autoRefreshCheck.SetChecked(true)

	controls := container.NewHBox(
		refreshBtn,
		autoRefreshCheck,
	)

	// Main layout
	content := container.NewBorder(
		container.NewVBox(statusCard, controls),
		nil,
		nil,
		nil,
		portCard,
	)

	// Start auto-refresh
	go d.startAutoRefresh()

	return content
}

// fetchHealth retrieves health data from the API
func (d *DashboardTab) fetchHealth() {
	resp, err := http.Get(d.apiURL + "/health")
	if err != nil {
		d.statusLabel.SetText("Status: Error - Cannot connect to service")
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		d.statusLabel.SetText("Status: Error - Cannot read response")
		return
	}

	var health HealthResponse
	if err := json.Unmarshal(body, &health); err != nil {
		d.statusLabel.SetText("Status: Error - Invalid response")
		return
	}

	// Update UI
	d.statusLabel.SetText(fmt.Sprintf("Status: %s", health.Status))
	d.instanceLabel.SetText(fmt.Sprintf("Instance: %s", health.InstanceID))
	d.versionLabel.SetText(fmt.Sprintf("Version: %s", health.Version))
	d.uptimeLabel.SetText(fmt.Sprintf("Uptime: %s", formatUptime(health.UptimeSec)))

	// Update port data
	d.portData = make([]PortInfo, 0, len(health.Ports))
	for _, port := range health.Ports {
		d.portData = append(d.portData, port)
	}

	// Refresh table
	d.portTable.Refresh()
}

// startAutoRefresh starts the automatic refresh loop
func (d *DashboardTab) startAutoRefresh() {
	ticker := time.NewTicker(d.refreshInterval)
	defer ticker.Stop()

	// Initial fetch
	d.fetchHealth()

	for {
		select {
		case <-ticker.C:
			d.fetchHealth()
		case <-d.stopRefresh:
			return
		}
	}
}

// formatUptime formats uptime seconds into a readable string
func formatUptime(seconds int64) string {
	duration := time.Duration(seconds) * time.Second
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	secs := int(duration.Seconds()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm %ds", hours, minutes, secs)
	} else if minutes > 0 {
		return fmt.Sprintf("%dm %ds", minutes, secs)
	}
	return fmt.Sprintf("%ds", secs)
}
