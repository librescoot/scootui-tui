package main

import (
	"fmt"
	"os"
	"os/exec"
	"scootui-tui/redis"

	tea "github.com/charmbracelet/bubbletea"
)

// Target display: 480x480 pixels with 8x16 console font = 60x30 chars
const (
	targetCols = 60
	targetRows = 30
)

func main() {
	redisHost := os.Getenv("SCOOTUI_REDIS_HOST")
	if redisHost == "" {
		redisHost = "192.168.7.1:6379"
	}

	takeOverFramebuffer()

	// Always create the client — go-redis reconnects automatically.
	// We don't bail on initial failure; the UI shows connection state.
	client := redis.NewClient(redisHost)

	m := NewModel(client)
	m.width = targetCols
	m.height = targetRows
	m.fixedSize = true

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}

// takeOverFramebuffer stops the boot-animation service and rebinds the
// framebuffer console so tty1 text renders to /dev/fb0. boot-animation
// unbinds vtcon1 in its ExecStartPre to keep the kernel cursor and printk
// off the Lottie frames; we reverse that here.
func takeOverFramebuffer() {
	_ = exec.Command("systemctl", "stop", "boot-animation.service").Run()
	const vtcon1Bind = "/sys/class/vtconsole/vtcon1/bind"
	if _, err := os.Stat(vtcon1Bind); err == nil {
		_ = os.WriteFile(vtcon1Bind, []byte("1\n"), 0)
	}
}
