package monitoring

import (
	"fmt"
	"net/http"

	"cdrgenerator/output"
)

// MetricsHandler creates an HTTP handler for Prometheus metrics
type MetricsHandler struct {
	manager *output.Manager
}

// NewMetricsHandler creates a new metrics handler
func NewMetricsHandler(manager *output.Manager) *MetricsHandler {
	return &MetricsHandler{
		manager: manager,
	}
}

// ServeHTTP handles the /metrics endpoint in Prometheus format
func (h *MetricsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	states := h.manager.GetChannelStates()

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	// Records total
	fmt.Fprintln(w, "# HELP cdrgenerator_records_total Total CDR records sent")
	fmt.Fprintln(w, "# TYPE cdrgenerator_records_total counter")
	for _, info := range states {
		fmt.Fprintf(w, "cdrgenerator_records_total{port=%q,format=%q,mode=%q} %d\n",
			info.Device, info.Format, info.Mode, info.RecordsSent)
	}

	// Bytes total
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "# HELP cdrgenerator_bytes_sent_total Total bytes sent")
	fmt.Fprintln(w, "# TYPE cdrgenerator_bytes_sent_total counter")
	for _, info := range states {
		fmt.Fprintf(w, "cdrgenerator_bytes_sent_total{port=%q} %d\n",
			info.Device, info.BytesSent)
	}

	// Errors total
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "# HELP cdrgenerator_port_errors_total Total port errors")
	fmt.Fprintln(w, "# TYPE cdrgenerator_port_errors_total counter")
	for _, info := range states {
		fmt.Fprintf(w, "cdrgenerator_port_errors_total{port=%q} %d\n",
			info.Device, info.Errors)
	}

	// Port status
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "# HELP cdrgenerator_port_up Port status (1=running, 0=not running)")
	fmt.Fprintln(w, "# TYPE cdrgenerator_port_up gauge")
	for _, info := range states {
		up := 0
		if info.State == "running" {
			up = 1
		}
		fmt.Fprintf(w, "cdrgenerator_port_up{port=%q,format=%q} %d\n",
			info.Device, info.Format, up)
	}

	// Last record timestamp
	fmt.Fprintln(w, "")
	fmt.Fprintln(w, "# HELP cdrgenerator_last_record_timestamp Unix timestamp of last record sent")
	fmt.Fprintln(w, "# TYPE cdrgenerator_last_record_timestamp gauge")
	for _, info := range states {
		if !info.LastRecordTime.IsZero() {
			fmt.Fprintf(w, "cdrgenerator_last_record_timestamp{port=%q} %d\n",
				info.Device, info.LastRecordTime.Unix())
		}
	}
}
