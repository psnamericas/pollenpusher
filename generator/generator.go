package generator

import (
	"context"
	"fmt"
	"os"
	"sync"

	"cdrgenerator/config"
	"cdrgenerator/format"
)

// Mode represents the generator mode
type Mode string

const (
	ModeReplay    Mode = "replay"
	ModeSynthetic Mode = "synthetic"
)

// Generator produces CDR records based on configuration
type Generator struct {
	format      format.CDRFormat
	mode        Mode
	portConfig  *config.PortConfig
	rateLimiter *RateLimiter
	genContext  *format.GenerationContext

	// For replay mode
	records      []format.CDRRecord
	recordIndex  int
	loop         bool
	recordsMutex sync.Mutex
}

// New creates a new generator for the given port configuration
func New(portCfg *config.PortConfig, jitterPercent float64) (*Generator, error) {
	// Get the format handler
	f, err := format.Get(portCfg.Format)
	if err != nil {
		return nil, fmt.Errorf("unknown format %s: %w", portCfg.Format, err)
	}

	mode := Mode(portCfg.Mode)
	if mode != ModeReplay && mode != ModeSynthetic {
		return nil, fmt.Errorf("invalid mode: %s", portCfg.Mode)
	}

	g := &Generator{
		format:      f,
		mode:        mode,
		portConfig:  portCfg,
		rateLimiter: NewRateLimiter(portCfg.CallsPerMinute, jitterPercent),
		loop:        portCfg.Loop,
	}

	// Initialize based on mode
	if mode == ModeReplay {
		if err := g.loadSampleFile(); err != nil {
			return nil, err
		}
	} else {
		// Synthetic mode - create generation context
		systemID := "default"
		psapName := "Default PSAP"
		if portCfg.Synthetic != nil {
			systemID = portCfg.Synthetic.SystemID
		}
		g.genContext = format.NewGenerationContext(systemID, psapName, 0)
	}

	return g, nil
}

// loadSampleFile loads and parses the sample file for replay mode
func (g *Generator) loadSampleFile() error {
	if g.portConfig.SampleFile == "" {
		return fmt.Errorf("sample_file is required for replay mode")
	}

	file, err := os.Open(g.portConfig.SampleFile)
	if err != nil {
		return fmt.Errorf("failed to open sample file: %w", err)
	}
	defer file.Close()

	records, err := g.format.ParseRecords(file)
	if err != nil {
		return fmt.Errorf("failed to parse sample file: %w", err)
	}

	if len(records) == 0 {
		return fmt.Errorf("no records found in sample file")
	}

	g.records = records
	g.recordIndex = 0
	return nil
}

// NextRecord returns the next CDR record
func (g *Generator) NextRecord(ctx context.Context) (*format.CDRRecord, error) {
	if g.mode == ModeReplay {
		return g.nextReplayRecord()
	}
	return g.nextSyntheticRecord()
}

// nextReplayRecord returns the next record from the sample file
func (g *Generator) nextReplayRecord() (*format.CDRRecord, error) {
	g.recordsMutex.Lock()
	defer g.recordsMutex.Unlock()

	if len(g.records) == 0 {
		return nil, fmt.Errorf("no records available")
	}

	record := g.records[g.recordIndex]
	g.recordIndex++

	// Handle looping
	if g.recordIndex >= len(g.records) {
		if g.loop {
			g.recordIndex = 0
		} else {
			return nil, fmt.Errorf("end of sample file reached")
		}
	}

	return &record, nil
}

// nextSyntheticRecord generates a new synthetic record
func (g *Generator) nextSyntheticRecord() (*format.CDRRecord, error) {
	if g.genContext == nil {
		return nil, fmt.Errorf("generation context not initialized")
	}

	return g.format.GenerateRecord(g.genContext)
}

// RateLimiter returns the rate limiter for this generator
func (g *Generator) RateLimiter() *RateLimiter {
	return g.rateLimiter
}

// Format returns the format handler
func (g *Generator) Format() format.CDRFormat {
	return g.format
}

// Mode returns the generator mode
func (g *Generator) Mode() Mode {
	return g.mode
}

// RecordCount returns the number of records (for replay mode)
func (g *Generator) RecordCount() int {
	if g.mode == ModeReplay {
		return len(g.records)
	}
	return -1 // Unlimited for synthetic
}
