package main

import (
	"fmt"
	"os"
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
