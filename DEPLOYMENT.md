# ScootUI TUI - Deployment Complete

## ✅ What's Installed

### Binary
- **Location**: `/data/scootui-tui`
- **Type**: Statically linked ARM binary (6.7MB)
- **Build**: ARMv7lhf with hard float, stripped

### Systemd Service
- **File**: `/etc/systemd/system/scootui-tui.service`
- **Status**: Installed and running
- **Output**: Runs on tty1 (framebuffer console)
- **Conflicts**: Automatically stops scootui (Flutter) and scootui-lvgl when started

### Switch Script
- **File**: `/data/switch-ui.sh`
- **Updated**: Now supports flutter, lvgl, and tui modes

## Usage

### Switch Between UIs
```bash
# Switch to TUI
/data/switch-ui.sh tui

# Switch to Flutter
/data/switch-ui.sh flutter

# Switch to LVGL
/data/switch-ui.sh lvgl

# Check status
/data/switch-ui.sh
```

### Direct Service Control
```bash
# Start TUI
systemctl start scootui-tui

# Stop TUI
systemctl stop scootui-tui

# Enable on boot
systemctl enable scootui-tui

# Check status
systemctl status scootui-tui

# View logs
journalctl -u scootui-tui -f
```

## Features

- **10Hz polling**: ECU and GPS data (smooth speedometer)
- **1Hz polling**: Battery, vehicle state, connectivity
- **Navigation**: Valhalla routing with turn-by-turn
- **Status bars**: Battery, time, GPS, Bluetooth, cellular
- **Trip stats**: Duration, average speed, distance
- **Debug mode**: Toggle with 'd' key (when running on console)

## Notes

- Output goes to tty1 (framebuffer), not visible via SSH
- Redis connection is optional (degrades gracefully if Redis is down)
- Auto-restarts on crash (RestartSec=5)
- Conflicts ensure only one UI runs at a time

## Build Info

Built with:
```bash
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 \
  go build -ldflags="-s -w" -o scootui-tui-arm
```

Checksum: 7af9290a7ee0b287539260e8818183bc
