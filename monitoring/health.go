package monitoring

import (
	"encoding/json"
	"net/http"
	"time"

	"cdrgenerator/output"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status     string                       `json:"status"`
	InstanceID string                       `json:"instance_id"`
	Version    string                       `json:"version"`
	UptimeSec  int64                        `json:"uptime_sec"`
	Ports      map[string]output.ChannelInfo `json:"ports"`
}

// HealthHandler creates an HTTP handler for health checks
type HealthHandler struct {
	instanceID string
	version    string
	startTime  time.Time
	manager    *output.Manager
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(instanceID, version string, manager *output.Manager) *HealthHandler {
	return &HealthHandler{
		instanceID: instanceID,
		version:    version,
		startTime:  time.Now(),
		manager:    manager,
	}
}

// ServeHTTP handles the /health endpoint
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	states := h.manager.GetChannelStates()

	// Determine overall status
	status := "healthy"
	for _, info := range states {
		if info.State == "error" || info.State == "reconnecting" {
			status = "degraded"
			break
		}
	}

	response := HealthResponse{
		Status:     status,
		InstanceID: h.instanceID,
		Version:    h.version,
		UptimeSec:  int64(time.Since(h.startTime).Seconds()),
		Ports:      states,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}
