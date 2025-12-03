package monitoring

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"cdrgenerator/config"
	"cdrgenerator/output"
)

//go:embed dashboard.html
var dashboardHTML string

// Server provides HTTP endpoints for monitoring
type Server struct {
	config  *config.MonitoringConfig
	manager *output.Manager
	server  *http.Server
	logger  *slog.Logger
}

// NewServer creates a new monitoring server
func NewServer(cfg *config.MonitoringConfig, instanceID, version string, manager *output.Manager, logger *slog.Logger) *Server {
	return NewServerWithConfigPath(cfg, instanceID, version, manager, logger, "/etc/cdrgenerator/config.json")
}

// NewServerWithConfigPath creates a new monitoring server with a custom config path
func NewServerWithConfigPath(cfg *config.MonitoringConfig, instanceID, version string, manager *output.Manager, logger *slog.Logger, configPath string) *Server {
	mux := http.NewServeMux()

	// Health endpoint
	healthHandler := NewHealthHandler(instanceID, version, manager)
	mux.Handle("/health", healthHandler)

	// Metrics endpoint (Prometheus format)
	metricsHandler := NewMetricsHandler(manager)
	mux.Handle("/metrics", metricsHandler)

	// Config endpoint
	configHandler := NewConfigHandler(configPath)
	mux.Handle("/api/config", configHandler)

	// Records endpoint
	recordsHandler := NewRecordsHandler(manager)
	mux.Handle("/api/records", recordsHandler)

	// System ports endpoint
	sysPortsHandler := NewSysPortsHandler()
	mux.Handle("/api/sysports", sysPortsHandler)

	// Dashboard endpoint
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, dashboardHTML)
	})

	return &Server{
		config:  cfg,
		manager: manager,
		server: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      mux,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// Start starts the monitoring server
func (s *Server) Start() error {
	s.logger.Info("Starting monitoring server", "port", s.config.Port)

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("Monitoring server error", "error", err)
		}
	}()

	return nil
}

// Stop gracefully stops the monitoring server
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping monitoring server")
	return s.server.Shutdown(ctx)
}
