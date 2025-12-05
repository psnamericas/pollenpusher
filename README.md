# PollenPusher 🌼

**Test CDR Data Generator for Emergency Services Systems**

PollenPusher generates realistic Call Detail Record (CDR) data for testing emergency services call handling systems. It simulates serial output from various 911 system formats including Vesta and Viper, making it ideal for testing data collection pipelines, analytics platforms, and reporting systems.

## Features

- **Multiple CDR Formats**: Supports Vesta, Viper, and other 911 system formats
- **Dual Output Modes**:
  - **Replay Mode**: Loop through real CDR sample files with realistic timing
  - **Synthetic Mode**: Generate fake but realistic CDR data on-the-fly
- **Configurable Call Rates**: Control calls-per-minute (CPM) for each output channel
- **Multi-Channel**: Support for multiple serial ports simultaneously
- **Realistic Timing**: Configurable jitter to simulate real-world variance
- **Web Dashboard**: Real-time monitoring at `http://localhost:8080`
- **System Monitoring**: View COM port status and activity

## Quick Start

### Prerequisites

- **Go 1.23.4+** installed
- **Serial ports** available (physical or virtual)
- **Linux/macOS** (Windows support via WSL)

### Installation

```bash
# Clone the repository
git clone https://github.com/psnamericas/pollenpusher.git
cd pollenpusher

# Build the binary
go build -o pollenpusher

# Copy example config
cp configs/example-config.json config.json

# Edit config to match your serial ports
nano config.json
```

### Basic Configuration

Edit `config.json` to configure your output channels:

```json
{
  "app": {
    "name": "PollenPusher",
    "instance_id": "test-01"
  },
  "ports": [
    {
      "device": "/dev/ttyS0",
      "baud_rate": 9600,
      "format": "vesta",
      "mode": "replay",
      "sample_file": "samples/Vesta/vestasample.csv",
      "loop": true,
      "calls_per_minute": 2.5,
      "enabled": true
    }
  ],
  "monitoring": {
    "port": 8080
  }
}
```

### Running

```bash
# Run with default config (config.json)
./pollenpusher

# Run with specific config
./pollenpusher -config /path/to/config.json

# Run in debug mode
./pollenpusher -debug
```

### Access the Dashboard

Open your browser to `http://localhost:8080` to view:
- Service status and uptime
- Active output channels
- Real-time record counts and byte statistics
- Recent records (click any channel row)
- System COM port information

## Configuration Guide

### Port Configuration

Each port in the `ports` array supports these options:

```json
{
  "device": "/dev/ttyS0",           // Serial device path
  "baud_rate": 9600,                // Baud rate (110-115200)
  "data_bits": 8,                   // Data bits (5-8)
  "stop_bits": 1,                   // Stop bits (1-2)
  "parity": "none",                 // none, odd, even, mark, space
  "format": "vesta",                // CDR format: vesta, viper
  "mode": "replay",                 // replay or synthetic
  "sample_file": "samples/...",     // Path to sample file (replay mode)
  "loop": true,                     // Loop sample file
  "calls_per_minute": 2.5,          // Target call rate
  "enabled": true,                  // Enable this port
  "description": "Test channel",    // Human-readable description

  // Synthetic mode only
  "synthetic": {
    "system_id": "PSAP-001",        // System identifier
    "agent_count": 15,               // Number of agents to simulate
    "min_duration_sec": 30,          // Min call duration
    "max_duration_sec": 600,         // Max call duration
    "include_agent_events": true    // Include agent login/logout
  }
}
```

### Timing Configuration

```json
{
  "timing": {
    "jitter_percent": 20,           // Timing variance (0-50%)
    "startup_delay_sec": 5          // Delay before first record
  }
}
```

### Monitoring Configuration

```json
{
  "monitoring": {
    "port": 8080,                   // Web dashboard port
    "stats_interval_sec": 60        // Stats update interval
  }
}
```

## Output Formats

### Vesta Format

Standard Vesta CDR format used by many 911 systems. Example:
```
01,12345,2024-12-04 10:30:45,555-1234,15,123 MAIN ST,ANYWHERE,TX,75001,FIRE
```

### Viper Format

Viper CDR format with agent events and extended metadata. Example:
```
CALL,001,2024-12-04T10:30:45Z,5551234567,AGENT001,120,COMPLETED
```

## Use Cases

### Testing Data Collection Pipelines

Connect PollenPusher to your data collector (e.g., NectarCollector) to test:
- Serial data ingestion
- Parser accuracy
- Data validation
- Error handling
- Performance under load

Example setup:
```
PollenPusher (COM1) → Serial Cable → NectarCollector (COM5)
                                    ↓
                              NATS / Database
```

### Load Testing

Generate high-volume CDR streams to test system capacity:

```json
{
  "ports": [
    {
      "device": "/dev/ttyS0",
      "mode": "synthetic",
      "calls_per_minute": 60,      // 1 call/second
      "enabled": true
    }
  ]
}
```

### Development & QA

Use replay mode to reproduce specific scenarios:

1. Capture real CDR samples from production
2. Save to `samples/` directory
3. Configure replay mode with the sample file
4. Reproduce issues in development environment

## API Endpoints

### Health Check
```bash
curl http://localhost:8080/health | jq
```

Response:
```json
{
  "status": "healthy",
  "instance_id": "test-01",
  "version": "1.0.0",
  "uptime_sec": 3600,
  "ports": {
    "/dev/ttyS0": {
      "device": "/dev/ttyS0",
      "state": "running",
      "format": "vesta",
      "mode": "replay",
      "records_sent": 150,
      "bytes_sent": 45000,
      "errors": 0,
      "last_record_time": "2024-12-04T10:45:23Z"
    }
  }
}
```

### System Ports
```bash
curl http://localhost:8080/api/sysports | jq
```

### Recent Records
```bash
curl http://localhost:8080/api/records?device=/dev/ttyS0 | jq
```

## Production Deployment

### Systemd Service

Create `/etc/systemd/system/pollenpusher.service`:

```ini
[Unit]
Description=PollenPusher CDR Generator
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/pollenpusher
ExecStart=/opt/pollenpusher/pollenpusher -config /etc/pollenpusher/config.json
Restart=on-failure
RestartSec=10s

[Install]
WantedBy=multi-user.target
```

Enable and start:
```bash
sudo systemctl daemon-reload
sudo systemctl enable pollenpusher
sudo systemctl start pollenpusher
sudo systemctl status pollenpusher
```

### Docker

```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o pollenpusher

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/pollenpusher /usr/local/bin/
COPY --from=builder /build/samples /samples
COPY --from=builder /build/configs/example-config.json /config.json
EXPOSE 8080
CMD ["/usr/local/bin/pollenpusher", "-config", "/config.json"]
```

Build and run:
```bash
docker build -t pollenpusher .
docker run -d -p 8080:8080 --device=/dev/ttyS0 pollenpusher
```

## Troubleshooting

### No Serial Ports Available

**Linux**: Create virtual serial ports using `socat`:
```bash
# Create a pair of virtual serial ports
socat -d -d pty,raw,echo=0 pty,raw,echo=0
# Returns something like:
# /dev/pts/2 <-> /dev/pts/3
```

**macOS**: Install virtual serial port driver or use USB-to-Serial adapters

### Permission Denied on Serial Port

```bash
# Add user to dialout group (Linux)
sudo usermod -a -G dialout $USER

# Or run with sudo
sudo ./pollenpusher
```

### Dashboard Not Accessible

Check if port 8080 is in use:
```bash
lsof -i :8080
# Or change port in config.json
```

### High CPU Usage

Reduce `calls_per_minute` or increase `timing.jitter_percent` to spread out load.

## Development

### Project Structure

```
pollenpusher/
├── main.go                 # Entry point
├── config/
│   ├── config.go          # Configuration types
│   └── validate.go        # Validation
├── generator/
│   ├── manager.go         # Channel manager
│   ├── channel.go         # Per-port generator
│   ├── replay.go          # Replay mode
│   └── synthetic.go       # Synthetic mode
├── formats/
│   ├── vesta.go           # Vesta format
│   └── viper.go           # Viper format
├── monitoring/
│   ├── server.go          # HTTP server
│   └── dashboard.html     # Web UI
├── samples/
│   ├── Vesta/             # Vesta samples
│   └── Viper/             # Viper samples
└── configs/
    └── example-config.json
```

### Adding New Formats

1. Create new file in `formats/` directory
2. Implement the `Generator` interface
3. Register in `generator/channel.go`
4. Add sample file to `samples/`

### Running Tests

```bash
go test ./...
```

### Building for Multiple Platforms

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o pollenpusher-linux

# macOS
GOOS=darwin GOARCH=amd64 go build -o pollenpusher-macos

# Windows
GOOS=windows GOARCH=amd64 go build -o pollenpusher.exe
```

## Performance

- **CPU**: < 1% typical
- **Memory**: ~20 MB
- **Throughput**: 100+ calls/minute per channel
- **Latency**: < 10ms record generation

## Architecture

PollenPusher uses a manager/channel pattern:

```
Manager
  ├─ Channel 1 (/dev/ttyS0) → Vesta Replay
  ├─ Channel 2 (/dev/ttyS1) → Viper Synthetic
  └─ Channel N (/dev/ttySN) → ...
```

Each channel runs in its own goroutine with:
- Independent timing control
- Separate error tracking
- Individual statistics
- Graceful shutdown support

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

MIT License - see LICENSE file for details

## Support

- **Issues**: https://github.com/psnamericas/pollenpusher/issues
- **Documentation**: https://github.com/psnamericas/pollenpusher/wiki

## Related Projects

- **NectarCollector**: Serial data collector with NATS streaming
- **HoneyView**: Real-time CDR monitoring dashboard

## Changelog

### v1.0.0 (2024-12-04)
- Initial release
- Vesta and Viper format support
- Replay and synthetic modes
- Web dashboard
- Multi-channel output

---

**Made with 🐝 for emergency services data testing**
