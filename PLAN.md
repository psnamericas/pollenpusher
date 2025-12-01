# CDRGenerator Implementation Plan

## Overview

A Go program that simulates 911 CDR (Call Detail Record) traffic over serial ports for testing data collector boxes deployed across Nebraska data centers. The generator runs on an industrial Linux Ubuntu server with 6 serial ports.

## Requirements

| Requirement | Detail |
|-------------|--------|
| Formats | Vesta and Viper initially, extensible for ECW, Solacom, etc. |
| Output Mode | Complete CDR blocks at "end of call" (not line-by-line) |
| Serial Config | Configurable baud rate per port |
| Data Source | Both replay (from sample CSV) and synthetic generation |
| Monitoring | HTTP health endpoint + Prometheus metrics |
| Notifications | Slack webhooks for startup/shutdown/errors |

---

## Project Structure

```
cdrgenerator/
├── main.go                    # Entry point, CLI, orchestration
├── go.mod
├── go.sum
├── config/
│   ├── config.go              # Configuration types and loading
│   └── validate.go            # Config validation
├── serial/
│   ├── port.go                # Serial port interface
│   ├── writer.go              # Implementation with stats
│   └── mock.go                # Mock for testing
├── format/
│   ├── format.go              # CDRFormat interface
│   ├── registry.go            # Format registration (plugin pattern)
│   ├── vesta/
│   │   ├── vesta.go           # Vesta format handler
│   │   ├── parser.go          # Parse sample CSV
│   │   └── generator.go       # Synthetic generation
│   └── viper/
│       ├── viper.go           # Viper format handler
│       ├── parser.go          # Parse sample CSV
│       └── generator.go       # Synthetic generation
├── generator/
│   ├── generator.go           # Record generation orchestration
│   ├── replay.go              # Replay mode (from samples)
│   ├── synthetic.go           # Synthetic mode
│   └── timing.go              # Rate limiting with jitter
├── output/
│   ├── channel.go             # Per-port output channel (goroutine)
│   └── manager.go             # Manages all output channels
├── monitoring/
│   ├── health.go              # HTTP /health endpoint
│   ├── metrics.go             # Prometheus metrics
│   └── stats.go               # Statistics collection
├── notify/
│   └── slack.go               # Slack webhook notifications
├── configs/
│   └── example-config.json    # Example configuration
├── samples/                   # Existing sample files
│   ├── Vesta/vestasample.csv
│   └── Viper/vipersample.csv
└── systemd/
    └── cdrgenerator.service   # systemd unit file
```

---

## Key Interfaces

### CDRFormat Interface

```go
// format/format.go
type CDRFormat interface {
    Name() string
    Description() string
    ParseRecords(reader io.Reader) ([]CDRRecord, error)
    GenerateRecord(ctx *GenerationContext) (CDRRecord, error)
    FormatOutput(record CDRRecord) []byte
}

type CDRRecord struct {
    ID        string
    Timestamp time.Time
    Duration  time.Duration
    Lines     []string  // The actual output lines
}

type GenerationContext struct {
    SystemID     string
    CurrentTime  time.Time
    Random       *rand.Rand
    AgentPool    []Agent
    LocationPool []Location
}
```

### Format Registry (Plugin Pattern)

```go
// format/registry.go
var registry = make(map[string]CDRFormat)

func Register(f CDRFormat) {
    registry[strings.ToLower(f.Name())] = f
}

func Get(name string) (CDRFormat, error) {
    f, ok := registry[strings.ToLower(name)]
    if !ok {
        return nil, fmt.Errorf("unknown format: %s", name)
    }
    return f, nil
}

func List() []string { ... }
```

Formats self-register via `init()`:

```go
// format/vesta/vesta.go
func init() {
    format.Register(&VestaFormat{})
}
```

Main imports formats for side effects:

```go
// main.go
import (
    _ "cdrgenerator/format/vesta"
    _ "cdrgenerator/format/viper"
)
```

---

## Configuration File Format

```json
{
  "app": {
    "name": "Nebraska CDR Generator",
    "instance_id": "ne-datacenter-01"
  },
  "ports": [
    {
      "device": "/dev/ttyUSB0",
      "baud_rate": 9600,
      "format": "vesta",
      "mode": "replay",
      "sample_file": "samples/Vesta/vestasample.csv",
      "loop": true,
      "calls_per_minute": 2.5,
      "enabled": true,
      "description": "Vesta simulator for Lincoln PSAP"
    },
    {
      "device": "/dev/ttyUSB1",
      "baud_rate": 19200,
      "format": "viper",
      "mode": "synthetic",
      "calls_per_minute": 1.0,
      "synthetic": {
        "system_id": "detroit",
        "agent_count": 10,
        "min_duration_sec": 30,
        "max_duration_sec": 300,
        "include_agent_events": true
      },
      "enabled": true,
      "description": "Viper simulator for Omaha PSAP"
    }
  ],
  "timing": {
    "jitter_percent": 20,
    "startup_delay_sec": 5
  },
  "logging": {
    "level": "info",
    "base_path": "/var/log/cdrgenerator/",
    "filename": "cdrgenerator.log",
    "max_size_mb": 50,
    "max_backups": 5,
    "compress": true
  },
  "monitoring": {
    "port": 8080,
    "stats_interval_sec": 60
  },
  "slack": {
    "webhook_url": "https://hooks.slack.com/services/...",
    "notify_startup": true,
    "notify_shutdown": true,
    "notify_errors": true
  },
  "recovery": {
    "reconnect_delay_sec": 5,
    "max_reconnect_delay_sec": 300,
    "exponential_backoff": true
  }
}
```

---

## CLI Interface

```
Usage: cdrgenerator [options]

Required:
  -config string     Path to configuration file

Optional:
  -validate          Validate configuration and exit (dry run)
  -list-ports        List available serial ports and exit
  -list-formats      List registered CDR formats and exit
  -debug             Enable debug logging (overrides config)
  -version           Display version information
```

---

## Data Format Details

### Vesta Format

Sample location: `samples/Vesta/vestasample.csv`

**Structure:**
- CSV with `sysident,message` columns
- Records separated by `---   ---   ---   ---   ---   ---` delimiter lines
- Data stored in descending sysident order (newest first)

**Record block contains (when outputting in order):**
1. PSAP identifier (e.g., "3001 Downriver")
2. Call events with timestamps (Arrives, Picks Up, Hangs Up, Finishes)
3. "ALI Information" marker
4. Location data with ANI, CPN, GPS coordinates, carrier info
5. "SIP Call IDs" marker
6. SIP identifiers
7. Separator line

**Example output block:**
```
3001 Downriver
ANI             7345118474                                                      CPN             3136955164                                                                                                                                      Call 10105964   Arrives On               DCDAEIM9112     Dec/01/25 15:55:19 EST ...
ALI Information
734-511-8474   CBN 313-695-5164    VZW  12/01/2025     15:56:24.0EST        WPH2VERIZON ...
SIP Call IDs
IRZPPIxPit68SUDfvx7UWw..
---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---   ---
```

### Viper Format

Sample location: `samples/Viper/vipersample.csv`

**Structure:**
- CSV with `sysident,message` columns
- CDR blocks: `===== CDR BEGIN : MM/DD/YY HH:MM:SS.mmm =====` to `===== CDR END =====`
- Agent blocks: `===== AGENT BEGIN : MM/DD/YY HH:MM:SS.mmm =====` to `===== AGENT END =====`

**CDR block contains:**
- System ID, Trunk Group info
- VoIP timing events: `HH:MM:SS.mmm [VoIP] Event description`
- TCI (Trunk Call Information) events
- ALI data with coordinates

**Agent block contains:**
- POS/STN (Position/Station) numbers
- Agent name, ID, role
- ACD queue assignments
- ON CALL / OFF CALL states

**Example CDR block:**
```
===== CDR BEGIN : 12/01/25 15:57:59.997 =====
00:00:00.000 [  TS] SYSTEM ID = detroit
00:00:00.000 [VoIP] Incoming Call(ID: 911221-39051-20251201205759) Offered on Trunk SIP001/1261630006-911221
00:00:00.000 [  TS] Trunk Group = 911
00:00:00.000 [VoIP] Call Presented
00:00:00.000 [VoIP] ANI: (40)'7345110366' [VALID] PseudoANI: '' [NONE]
00:00:00.104 [VoIP] Call Connected
00:00:00.108 [VoIP] Routing call QUEUE = 6001
00:00:09.322 [VoIP] Call Terminated
00:00:09.322 [  TS] Call Completed
===== CDR END =====
```

---

## Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Initialize Go module and dependencies
- [ ] `config/config.go` - Configuration struct definitions
- [ ] `config/validate.go` - Configuration validation
- [ ] `format/format.go` - CDRFormat interface and CDRRecord type
- [ ] `format/registry.go` - Format registration
- [ ] `serial/port.go` - Serial port interface
- [ ] `serial/writer.go` - Serial port implementation
- [ ] `serial/mock.go` - Mock writer for testing
- [ ] `main.go` - Basic CLI parsing and startup

### Phase 2: Vesta Format
- [ ] `format/vesta/parser.go` - Parse vestasample.csv into CDRRecord blocks
- [ ] `format/vesta/generator.go` - Synthetic Vesta record generation
- [ ] `format/vesta/vesta.go` - CDRFormat implementation
- [ ] Unit tests for Vesta parsing and generation

### Phase 3: Viper Format
- [ ] `format/viper/parser.go` - Parse vipersample.csv (CDR/AGENT blocks)
- [ ] `format/viper/generator.go` - Synthetic Viper record generation
- [ ] `format/viper/viper.go` - CDRFormat implementation
- [ ] Unit tests for Viper parsing and generation

### Phase 4: Generator and Output
- [ ] `generator/timing.go` - Rate limiting with calls_per_minute and jitter
- [ ] `generator/replay.go` - Replay mode from parsed sample records
- [ ] `generator/synthetic.go` - Synthetic mode using format generators
- [ ] `output/channel.go` - Per-port goroutine with reconnection logic
- [ ] `output/manager.go` - Start/stop all channels, graceful shutdown

### Phase 5: Monitoring and Notifications
- [ ] `monitoring/stats.go` - Per-port statistics collection
- [ ] `monitoring/health.go` - HTTP /health endpoint
- [ ] `monitoring/metrics.go` - Prometheus /metrics endpoint
- [ ] `notify/slack.go` - Slack webhook integration

### Phase 6: Production Readiness
- [ ] Signal handling (SIGTERM, SIGINT) with graceful shutdown
- [ ] `systemd/cdrgenerator.service` - Systemd unit file
- [ ] `configs/example-config.json` - Example configuration
- [ ] Log rotation with lumberjack
- [ ] README.md with usage instructions

---

## Monitoring Endpoints

### Health Check (GET /health)

```json
{
  "status": "healthy",
  "uptime_sec": 86400,
  "instance_id": "ne-datacenter-01",
  "ports": {
    "/dev/ttyUSB0": {
      "status": "running",
      "format": "vesta",
      "mode": "replay",
      "records_sent": 12345,
      "bytes_sent": 5678901,
      "last_record_time": "2025-12-01T15:30:00Z",
      "errors_total": 2
    },
    "/dev/ttyUSB1": {
      "status": "reconnecting",
      "format": "viper",
      "mode": "synthetic",
      "records_sent": 5432,
      "bytes_sent": 2345678,
      "last_record_time": "2025-12-01T15:29:45Z",
      "errors_total": 5,
      "last_error": "device disconnected"
    }
  }
}
```

### Prometheus Metrics (GET /metrics)

```
# HELP cdrgenerator_records_total Total CDR records sent
# TYPE cdrgenerator_records_total counter
cdrgenerator_records_total{port="/dev/ttyUSB0",format="vesta"} 12345
cdrgenerator_records_total{port="/dev/ttyUSB1",format="viper"} 5432

# HELP cdrgenerator_bytes_sent_total Total bytes sent
# TYPE cdrgenerator_bytes_sent_total counter
cdrgenerator_bytes_sent_total{port="/dev/ttyUSB0"} 5678901

# HELP cdrgenerator_port_errors_total Total port errors
# TYPE cdrgenerator_port_errors_total counter
cdrgenerator_port_errors_total{port="/dev/ttyUSB0"} 2

# HELP cdrgenerator_port_up Port status (1=up, 0=down)
# TYPE cdrgenerator_port_up gauge
cdrgenerator_port_up{port="/dev/ttyUSB0"} 1
```

---

## Dependencies

```go
// go.mod
module cdrgenerator

go 1.21

require (
    go.bug.st/serial v1.6.1                  // Cross-platform serial ports
    gopkg.in/natefinch/lumberjack.v2 v2.2.1  // Log rotation
    github.com/prometheus/client_golang v1.18.0  // Prometheus metrics
)
```

---

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| One goroutine per port | Isolation ensures one port failure doesn't affect others |
| Format self-registration via init() | New formats added without modifying core code |
| Block-based output | Complete CDR records sent as units (matches real system behavior) |
| Pre-load sample files | Parse at startup, not during operation for consistent timing |
| Exponential backoff for reconnection | Prevents hammering failed ports |
| Context-based cancellation | Clean shutdown propagation through all goroutines |

---

## Adding a New Format (e.g., ECW)

1. Create directory: `format/ecw/`

2. Implement the format:
```go
// format/ecw/ecw.go
package ecw

import "cdrgenerator/format"

func init() {
    format.Register(&ECWFormat{})
}

type ECWFormat struct{}

func (f *ECWFormat) Name() string { return "ecw" }
func (f *ECWFormat) Description() string { return "Emergency CallWorks" }
func (f *ECWFormat) ParseRecords(r io.Reader) ([]format.CDRRecord, error) { ... }
func (f *ECWFormat) GenerateRecord(ctx *format.GenerationContext) (format.CDRRecord, error) { ... }
func (f *ECWFormat) FormatOutput(record format.CDRRecord) []byte { ... }
```

3. Import in main.go:
```go
import (
    _ "cdrgenerator/format/vesta"
    _ "cdrgenerator/format/viper"
    _ "cdrgenerator/format/ecw"  // New format
)
```

4. Use in config:
```json
{
  "device": "/dev/ttyUSB2",
  "format": "ecw",
  ...
}
```

---

## Systemd Service

```ini
# /etc/systemd/system/cdrgenerator.service
[Unit]
Description=CDR Generator for 911 Testing
After=network.target

[Service]
Type=simple
User=cdrgenerator
Group=dialout
ExecStart=/usr/local/bin/cdrgenerator -config /etc/cdrgenerator/config.json
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
```

Installation:
```bash
sudo cp cdrgenerator /usr/local/bin/
sudo cp configs/config.json /etc/cdrgenerator/
sudo cp systemd/cdrgenerator.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable cdrgenerator
sudo systemctl start cdrgenerator
```
