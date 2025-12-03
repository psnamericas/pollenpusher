package monitoring

import (
	"encoding/json"
	"net/http"

	"cdrgenerator/output"
)

// RecordsHandler handles requests for recent CDR records
type RecordsHandler struct {
	manager *output.Manager
}

// NewRecordsHandler creates a new records handler
func NewRecordsHandler(manager *output.Manager) *RecordsHandler {
	return &RecordsHandler{
		manager: manager,
	}
}

// ServeHTTP handles record requests
func (h *RecordsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	device := r.URL.Query().Get("device")
	if device == "" {
		http.Error(w, "device parameter required", http.StatusBadRequest)
		return
	}

	records := h.manager.GetRecentRecords(device, 10)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"device":  device,
		"records": records,
	})
}
