# ScootUI TUI

A terminal-based (TUI) dashboard for Librescoot electric scooters. Displays real-time vehicle telemetry, navigation, and status information in a text interface optimized for the 480x480 framebuffer console on iMX6DL hardware.

Part of the [Librescoot](https://librescoot.org/) open-source platform.

## Features

- **Real-time Vehicle Data**: Speed, battery status, motor power, GPS position
- **Status Indicators**: Blinkers, warnings, connectivity (GPS, Bluetooth, cellular)
- **Trip Statistics**: Duration, average speed, distance, odometer
- **Turn-by-Turn Navigation**: Valhalla routing integration with upcoming turns list
- **Optimized Polling**: 10Hz for ECU/GPS, 1Hz for other data
- **Debug Mode**: Toggle detailed system information display

## Requirements

- iMX6DL (ARMv7lhf) or compatible ARM device
- Redis server (192.168.7.1:6379 by default)
- Valhalla routing service (optional, for navigation)
- Framebuffer console (`/dev/fb0`)

## Building

### Local (development)
```bash
cd scootui-tui
go build -o scootui-tui
```

### ARM (production for iMX6DL)
```bash
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -ldflags="-s -w" -o scootui-tui-arm
```

## Installation on DBC

### Deploy binary
```bash
# Transfer to DBC /data directory
scp -J deep-blue scootui-tui-arm root@192.168.7.2:/data/scootui-tui
ssh -J deep-blue root@192.168.7.2 "chmod +x /data/scootui-tui"
```

### Install systemd service (optional)
```bash
# Copy service file and enable
scp -J deep-blue scootui-tui.service root@192.168.7.2:/etc/systemd/system/
ssh -J deep-blue root@192.168.7.2 "systemctl daemon-reload && systemctl enable scootui-tui"
```

## Running

### Manual
```bash
# SSH to DBC
ssh -J deep-blue root@192.168.7.2

# Run from /data
/data/scootui-tui

# Or with custom Redis host
SCOOTUI_REDIS_HOST=192.168.7.1:6379 /data/scootui-tui
```

### As systemd service
```bash
systemctl start scootui-tui
systemctl status scootui-tui
journalctl -u scootui-tui -f
```

## Configuration

Environment variables:

- `SCOOTUI_REDIS_HOST` - Redis server address (default: `192.168.7.1:6379`)
- `VALHALLA_ENDPOINT` - Valhalla routing service URL (default: `http://localhost:8002`)
- `SCOOTUI_LAYOUT` - Layout mode: `full`, `compact`, `auto` (default: `auto`)
- `SCOOTUI_THEME` - Color theme: `auto`, `light`, `dark` (default: `auto`)
- `SCOOTUI_DEBUG` - Enable debug mode on start: `true`, `false` (default: `false`)
- `TERM` - Terminal type (default: `linux`)

## Keyboard Controls

- `q` or `Ctrl+C` - Quit application
- `r` - Force refresh all data
- `d` - Toggle debug mode

## Architecture

### Data Flow
- **10Hz (100ms)**: ECU (speed, motor data) and GPS updates for smooth speedometer
- **1Hz (1s)**: Battery, vehicle state, connectivity, navigation, speed limit
- **2Hz (500ms)**: Blinker flash animation

### Redis Channels
Subscribes to same channels as Flutter scootui:
- `engine-ecu` - Motor speed, voltage, current, power
- `gps` - Location, speed, course
- `vehicle` - Blinkers, brakes, kickstand, state
- `battery:0`, `battery:1` - Dual battery packs
- `navigation` - Destination and route
- `internet` - Cellular connectivity
- `ble` - Bluetooth status
- `speed-limit` - Current speed limit

### Navigation
- Automatically calculates route when destination is set in Redis
- Uses Valhalla routing service (`/route` endpoint)
- Displays turn-by-turn instructions with distance
- Updates current instruction based on GPS position

## Development

### Testing with Simulator
Use the existing Python simulator scripts from scootui:

```bash
# Terminal 1: Start Redis (if not running)
redis-server

# Terminal 2: Simulate GPS data
cd scootui
python3 simulate-gps.py 37.7749 -122.4194

# Terminal 3: Simulate ride
python3 simulate-ride.py 37.7749 -122.4194

# Terminal 4: Run TUI
cd ../scootui-tui
SCOOTUI_REDIS_HOST=localhost:6379 ./scootui-tui
```

### Testing Navigation
```bash
# Simulate route following
cd scootui
python3 simulate-route-following.py 37.7749 -122.4194 37.8044 -122.2712

# Or manually set destination in Redis
redis-cli HSET navigation latitude "37.8044"
redis-cli HSET navigation longitude "-122.2712"
redis-cli HSET navigation address "Destination Address"
redis-cli HSET navigation timestamp "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

## Troubleshooting

### Redis Connection Failed
Check Redis is running and accessible:
```bash
redis-cli -h 192.168.7.1 -p 6379 PING
```

### No GPS Data
Check GPS state in Redis:
```bash
redis-cli HGET gps state
redis-cli HGETALL gps
```

### Valhalla Route Errors
Verify Valhalla service is running:
```bash
curl http://localhost:8002/status
```

### Framebuffer Issues
Check framebuffer device:
```bash
ls -l /dev/fb0
fbset -i
```

## License

This project is dual-licensed. The source code is available under the
[Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License][cc-by-nc-sa].
The maintainers reserve the right to grant separate licenses for commercial distribution; please contact the maintainers to discuss commercial licensing.

[![CC BY-NC-SA 4.0][cc-by-nc-sa-image]][cc-by-nc-sa]

[cc-by-nc-sa]: http://creativecommons.org/licenses/by-nc-sa/4.0/
[cc-by-nc-sa-image]: https://licensebuttons.net/l/by-nc-sa/4.0/88x31.png
