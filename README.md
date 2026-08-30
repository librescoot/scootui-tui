# Librescoot ScootUI TUI

Part of the [Librescoot](https://librescoot.org/) open-source platform.

## Overview

ScootUI TUI is a terminal dashboard for Librescoot vehicles. It renders vehicle
state in a Bubble Tea terminal application sized for the 480 × 480 framebuffer
console used by the DBC, while remaining usable in a regular terminal for
development.

## Capabilities

- Displays vehicle, motor, battery, GPS, connectivity, dashboard, OTA, and
  speed-limit state from the Redis-compatible datastore.
- Provides cluster, navigation, settings, and about screens.
- Calculates and presents turn-by-turn routes when a destination and a recent
  GPS fix are available.
- Handles dashboard button gestures and keyboard navigation.
- Continues running when the datastore is unavailable and reports connection
  state as it reconnects.

## Operation and interfaces

The application reads datastore hashes such as `vehicle`, `engine-ecu`,
`battery:0`, `battery:1`, `gps`, `navigation`, `settings`, `ota`, `internet`,
`ble`, and `speed-limit`. It uses `dashboard.valhalla-url` from `settings` as
its routing endpoint when present.

Keyboard controls are intended for development and console use:

| Key | Action |
| --- | --- |
| `q`, `Ctrl+C` | Quit |
| `Tab`, `1`–`4` | Select a main screen |
| `r` | Refresh data on cluster or navigation screens |
| `d` | Toggle cluster debug information |
| Arrow keys, `j`, `k`, `Enter`, `Esc` | Navigate settings and about screens |

On the target, the program stops `boot-animation.service` and rebinds the
framebuffer console before opening its alternate screen.

## Configuration

`SCOOTUI_REDIS_HOST` is the only application environment variable. It accepts a
`host:port` endpoint and defaults to `192.168.7.1:6379`.

Runtime dashboard preferences are read from the `settings` hash. The TUI can
write settings through its menu, but it does not publish a settings notification;
use the platform's normal settings-management path when other services must
react immediately.

## Build and test

The module declares Go 1.21. Build and check it with standard Go commands:

```sh
go build -o scootui-tui .
go test ./...
```

For the ARMv7 target, use:

```sh
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 \
  go build -ldflags='-s -w' -o scootui-tui .
```

There are currently no repository test files; `go test ./...` remains useful as
a package build and test discovery check.

## Deployment and runtime dependencies

The Yocto recipe installs `/usr/bin/scootui-tui` and a disabled-by-default
`scootui-tui.service`. The unit owns `/dev/tty1`, sets `TERM=linux`, conflicts
with `scootui.service` and `scootui-qt.service`, and requires a console device.
Enable exactly the UI service intended for the target:

```sh
systemctl enable --now scootui-tui.service
journalctl -u scootui-tui.service -f
```

A target deployment needs systemd, a framebuffer-backed Linux console, and a
Redis-compatible datastore. Navigation additionally needs a reachable
Valhalla-compatible routing endpoint.

## Operational notes

The service runs as root because it controls the boot-animation service and
console binding. Starting it manually on a non-target system can stop a local
`boot-animation.service` if one exists. It is a display client, not a source of
vehicle truth: it degrades when datastore data or routing is unavailable.

## License

This project is licensed under the [Creative Commons Attribution-NonCommercial-ShareAlike 4.0 International License](LICENSE).

Made with ❤️ by the Librescoot community
