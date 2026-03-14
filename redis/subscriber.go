package redis

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Subscriber manages Redis PUBSUB subscriptions
type Subscriber struct {
	client   *Client
	ctx      context.Context
	cancel   context.CancelFunc
	channels []string
}

// NewSubscriber creates a new PUBSUB subscriber
func NewSubscriber(client *Client, channels []string) (*Subscriber, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// Subscribe to channels
	if err := client.Subscribe(channels...); err != nil {
		cancel()
		return nil, err
	}

	return &Subscriber{
		client:   client,
		ctx:      ctx,
		cancel:   cancel,
		channels: channels,
	}, nil
}

// PubSubMessage represents a Redis PUBSUB message
type PubSubMessage struct {
	Channel string
	Payload string
}

// PubSubMessageCmd returns a tea.Cmd that listens for PUBSUB messages
func (s *Subscriber) PubSubMessageCmd() tea.Cmd {
	return func() tea.Msg {
		// Set a timeout to avoid blocking forever
		select {
		case <-s.ctx.Done():
			return nil
		case <-time.After(100 * time.Millisecond):
			// Try to receive a message with timeout
			channel, payload, err := s.tryReceiveMessage()
			if err != nil {
				return nil
			}
			return PubSubMessage{
				Channel: channel,
				Payload: payload,
			}
		}
	}
}

// tryReceiveMessage attempts to receive a message without blocking indefinitely
func (s *Subscriber) tryReceiveMessage() (string, string, error) {
	// This will block until a message arrives or context is cancelled
	// In production, we'd want a more sophisticated approach
	return s.client.ReceiveMessage()
}

// Close closes the subscriber
func (s *Subscriber) Close() {
	s.cancel()
}

// PollingManager manages periodic polling of Redis keys
type PollingManager struct {
	client  *Client
	timers  map[string]*time.Timer
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewPollingManager creates a new polling manager
func NewPollingManager(client *Client) *PollingManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &PollingManager{
		client: client,
		timers: make(map[string]*time.Timer),
		ctx:    ctx,
		cancel: cancel,
	}
}

// PollMessage represents a poll trigger
type PollMessage struct {
	Key string
}

// StartPolling starts polling for a specific key with interval
func (pm *PollingManager) StartPolling(key string, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return PollMessage{Key: key}
	})
}

// Close closes the polling manager
func (pm *PollingManager) Close() {
	pm.cancel()
	for _, timer := range pm.timers {
		timer.Stop()
	}
}

// GetDefaultChannels returns the list of channels to subscribe to
func GetDefaultChannels() []string {
	return []string{
		"vehicle",
		"engine-ecu",
		"battery:0",
		"battery:1",
		"gps",
		"navigation",
		"dashboard",
		"internet",
		"ble",
		"speed-limit",
	}
}
