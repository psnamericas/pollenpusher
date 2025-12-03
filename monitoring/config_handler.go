package monitoring

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"cdrgenerator/config"
	"cdrgenerator/format"
)

// ConfigHandler handles configuration management requests
type ConfigHandler struct {
	configPath string
}

// NewConfigHandler creates a new config handler
func NewConfigHandler(configPath string) *ConfigHandler {
	return &ConfigHandler{
		configPath: configPath,
	}
}

// ServeHTTP handles config requests
func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case http.MethodGet:
		h.getConfig(w, r)
	case http.MethodPost:
		h.saveConfig(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *ConfigHandler) getConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load(h.configPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(cfg)
}

func (h *ConfigHandler) saveConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cfg config.Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate configuration
	if err := config.Validate(&cfg, format.List()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Save to file
	data, err := json.MarshalIndent(&cfg, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := os.WriteFile(h.configPath, data, 0644); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Configuration saved. Restart the service to apply changes.",
	})
}
