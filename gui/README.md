# CDR Generator GUI

A graphical user interface for managing and monitoring the CDR Generator service.

## Features

### Dashboard Tab
- Real-time monitoring of CDR generator service
- View service status (running/stopped/failed)
- Monitor all configured ports:
  - Device name
  - CDR format (vesta/viper)
  - State (running/error)
  - Records sent
  - Bytes transmitted
  - Error count
  - Last record timestamp
- Auto-refresh every 2 seconds
- Manual refresh option

### Port Configuration Tab
- View all configured serial ports
- Add new port configurations
- Edit existing port settings:
  - Device path (/dev/ttyS0, etc.)
  - Baud rate
  - CDR format
  - Mode (replay/synthetic)
  - Sample file path
  - Calls per minute rate
  - Enable/disable individual ports
  - Description
- Delete port configurations
- Save configuration to file
- Reload configuration from disk

### Service Control Tab
- Start the CDR generator service
- Stop the CDR generator service
- Restart the service
- Check service status
- Enable/disable auto-start on boot
- View command output logs

## Requirements

- Go 1.21 or later
- CDR Generator service installed
- systemd (for service control)
- X11 or Wayland display server (Linux)
- macOS 10.12 or later (macOS)

## Building

```bash
cd gui
go build -o cdrgenerator-gui
```

## Running

### On macOS (Development)
```bash
./cdrgenerator-gui
```

### On Linux (Production)
```bash
# Run with sudo for systemctl commands
sudo ./cdrgenerator-gui
```

## Installation on Production Server

1. Build for Linux on your Mac:
```bash
# Install cross-compilation tools
go install github.com/fyne-io/fyne-cross@latest

# Build for Linux
cd gui
fyne-cross linux -arch amd64

# Binary will be in fyne-cross/dist/linux-amd64/
```

2. Copy to server:
```bash
scp fyne-cross/dist/linux-amd64/cdrgenerator-gui root@100.88.226.119:/usr/local/bin/
ssh root@100.88.226.119 "chmod +x /usr/local/bin/cdrgenerator-gui"
```

3. Run on server (requires X11 forwarding or VNC):
```bash
# With X11 forwarding
ssh -X root@100.88.226.119
cdrgenerator-gui

# Or with VNC/remote desktop
```

## Configuration

The GUI looks for the configuration file in the following locations:
1. `/etc/cdrgenerator/config.json` (production)
2. `configs/example-config.json` (development fallback)

## API Connection

The GUI connects to the CDR generator monitoring API at:
- URL: `http://localhost:8080`
- Endpoint: `/health`

Ensure the CDR generator service is running with the monitoring server enabled.

## Troubleshooting

### "Cannot connect to service"
- Ensure the CDR generator service is running
- Check that port 8080 is accessible
- Verify firewall settings

### "Failed to load config"
- Check that the config file exists
- Verify file permissions
- Ensure JSON syntax is valid

### Service control doesn't work
- Run the GUI with sudo: `sudo ./cdrgenerator-gui`
- Verify systemd service is installed
- Check service name matches: `cdrgenerator.service`

## Development

The GUI is built using [Fyne](https://fyne.io/), a cross-platform GUI toolkit for Go.

Directory structure:
```
gui/
├── main.go           # Application entry point
└── ui/
    ├── main.go       # Main UI layout and tabs
    ├── dashboard.go  # Dashboard tab implementation
    ├── portconfig.go # Port configuration tab
    └── control.go    # Service control tab
```

## License

Same as CDR Generator project.
