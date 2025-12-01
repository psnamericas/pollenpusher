package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cdrgenerator/config"
	"cdrgenerator/format"
	"cdrgenerator/serial"

	// Import format packages for side-effect registration
	_ "cdrgenerator/format/vesta"
	_ "cdrgenerator/format/viper"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	version   = "1.0.0"
	buildTime = "unknown"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to configuration file (required)")
	validate := flag.Bool("validate", false, "Validate configuration and exit")
	listPorts := flag.Bool("list-ports", false, "List available serial ports and exit")
	listFormats := flag.Bool("list-formats", false, "List registered CDR formats and exit")
	debug := flag.Bool("debug", false, "Enable debug logging")
	showVersion := flag.Bool("version", false, "Display version information")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "CDRGenerator - 911 CDR Traffic Simulator\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s -config config.json\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -config config.json -validate\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -list-formats\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("CDRGenerator version %s (built %s)\n", version, buildTime)
		os.Exit(0)
	}

	// Handle list-ports flag
	if *listPorts {
		ports, err := serial.ListPorts()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing ports: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Available serial ports:")
		if len(ports) == 0 {
			fmt.Println("  (none found)")
		} else {
			for _, port := range ports {
				fmt.Printf("  %s\n", port)
			}
		}
		os.Exit(0)
	}

	// Handle list-formats flag
	if *listFormats {
		fmt.Println("Registered CDR formats:")
		formats := format.List()
		if len(formats) == 0 {
			fmt.Println("  (none registered)")
		} else {
			format.ForEach(func(name string, f format.CDRFormat) {
				fmt.Printf("  %-10s - %s\n", name, f.Description())
			})
		}
		os.Exit(0)
	}

	// Require config path for main operation
	if *configPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -config flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(cfg, format.List()); err != nil {
		fmt.Fprintf(os.Stderr, "Configuration validation failed:\n  %v\n", err)
		os.Exit(1)
	}

	// Handle validate flag
	if *validate {
		fmt.Println("Configuration is valid")
		fmt.Printf("  Instance: %s\n", cfg.App.InstanceID)
		fmt.Printf("  Ports configured: %d\n", len(cfg.Ports))
		for i, port := range cfg.Ports {
			if port.Enabled {
				fmt.Printf("    [%d] %s - %s mode, %s format, %d baud\n",
					i, port.Device, port.Mode, port.Format, port.BaudRate)
			}
		}
		os.Exit(0)
	}

	// Setup logging
	logger := setupLogging(cfg, *debug)
	slog.SetDefault(logger)

	logger.Info("CDRGenerator starting",
		"version", version,
		"instance", cfg.App.InstanceID,
		"ports", len(cfg.Ports),
	)

	// Create context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal", "signal", sig)
		cancel()
	}()

	// TODO: Start output manager and monitoring server
	// For now, just wait for shutdown
	logger.Info("CDRGenerator ready - waiting for implementation of output channels")

	<-ctx.Done()
	logger.Info("CDRGenerator shutting down")
}

func setupLogging(cfg *config.Config, debug bool) *slog.Logger {
	level := slog.LevelInfo
	if debug {
		level = slog.LevelDebug
	} else {
		switch cfg.Logging.Level {
		case "debug":
			level = slog.LevelDebug
		case "warn":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler

	// If base path is set, use file logging with rotation
	if cfg.Logging.BasePath != "" {
		logPath := filepath.Join(cfg.Logging.BasePath, cfg.Logging.Filename)
		writer := &lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    cfg.Logging.MaxSizeMB,
			MaxBackups: cfg.Logging.MaxBackups,
			Compress:   cfg.Logging.Compress,
		}
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		// Use console logging
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}
