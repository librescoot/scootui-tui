package main

import (
	"fmt"
	"scootui-tui/components"
	"scootui-tui/input"
	"scootui-tui/redis"
	"scootui-tui/valhalla"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Screen identifies the active view.
type Screen int

const (
	ScreenCluster    Screen = iota
	ScreenNavigation
	ScreenSettings // overlay, only when parked
	ScreenAbout    // overlay, entered from settings menu
	numMainScreens = 2 // only Cluster and Navigation cycle
)

// Model holds application state
type Model struct {
	redisClient *redis.Client
	width       int
	height      int

	// Active screen
	activeScreen Screen

	// Vehicle data
	vehicle    *redis.VehicleData
	engine     *redis.EngineData
	battery0   *redis.BatteryData
	battery1   *redis.BatteryData
	gps        *redis.GpsData
	navigation *redis.NavigationData
	dashboard  *redis.DashboardData
	internet   *redis.InternetData
	bluetooth  *redis.BluetoothData
	speedLimit *redis.SpeedLimitData
	trip       *redis.TripData
	settings   *redis.SettingsData
	ota        *redis.OtaData

	// UI state
	blinkerFlash bool
	debugMode    bool
	quitting     bool

	// Navigation
	valhallaClient *valhalla.Client
	route          *valhalla.Route
	routeError     string

	// Menu state (for settings screen)
	menuState components.MenuState

	// About screen scroll
	aboutScroll int

	// Physical input gesture detection
	gestures *input.GestureDetector

	// Toast notifications
	toasts components.ToastManager

	// When true, ignore WindowSizeMsg and keep fixed width/height
	fixedSize bool
}

// NewModel creates initial model
func NewModel(redisClient *redis.Client) Model {
	return Model{
		redisClient:    redisClient,
		vehicle:        &redis.VehicleData{},
		engine:         &redis.EngineData{},
		battery0:       &redis.BatteryData{ID: 0},
		battery1:       &redis.BatteryData{ID: 1},
		gps:            &redis.GpsData{},
		navigation:     &redis.NavigationData{},
		dashboard:      &redis.DashboardData{},
		internet:       &redis.InternetData{},
		bluetooth:      &redis.BluetoothData{},
		speedLimit:     &redis.SpeedLimitData{},
		trip:           &redis.TripData{},
		settings:       redis.DefaultSettings(),
		ota:            &redis.OtaData{},
		valhallaClient: valhalla.NewClient(),
		menuState:      buildSettingsMenu(redis.DefaultSettings()),
		gestures:       input.NewGestureDetector(),
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fastTickCmd(),
		slowTickCmd(),
		blinkerTickCmd(),
		fetchAllDataCmd(),
		m.listenButtons(),
	)
}

// listenButtons subscribes to the buttons PUBSUB channel and forwards events.
func (m Model) listenButtons() tea.Cmd {
	return func() tea.Msg {
		sub := m.redisClient.SubscribeChannel("buttons")
		if sub == nil {
			// Subscription failed (Redis down) — retry after a delay
			time.Sleep(2 * time.Second)
			return buttonRetryMsg{}
		}

		msg := sub.ReceiveMessage()
		if msg == "" {
			return buttonRetryMsg{}
		}

		ctrl, pressed, ok := input.ParseButtonEvent(msg)
		if !ok {
			return buttonRetryMsg{}
		}
		return buttonEventMsg{control: ctrl, pressed: pressed}
	}
}

type buttonRetryMsg struct{}

// Messages

type tickMsg time.Time
type slowTickMsg time.Time
type blinkerTickMsg time.Time
type fetchDataMsg struct{}
type fetchSlowDataMsg struct{}
type dataUpdateMsg struct {
	vehicle    *redis.VehicleData
	engine     *redis.EngineData
	battery0   *redis.BatteryData
	battery1   *redis.BatteryData
	gps        *redis.GpsData
	navigation *redis.NavigationData
	dashboard  *redis.DashboardData
	internet   *redis.InternetData
	bluetooth  *redis.BluetoothData
	speedLimit *redis.SpeedLimitData
	settings   *redis.SettingsData
	ota        *redis.OtaData
	err        error
}

type routeCalculatedMsg struct {
	route *valhalla.Route
	err   error
}

type buttonEventMsg struct {
	control input.Control
	pressed bool
}

type gestureMsg input.GestureEvent

// Commands

func fastTickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func slowTickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return slowTickMsg(t)
	})
}

func blinkerTickCmd() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return blinkerTickMsg(t)
	})
}

func fetchFastData(m Model) tea.Msg {
	engine, err := m.redisClient.GetEngineData()
	if err != nil {
		return dataUpdateMsg{err: err}
	}
	gps, _ := m.redisClient.GetGpsData()

	return dataUpdateMsg{
		engine: engine,
		gps:    gps,
	}
}

func fetchSlowData(m Model) tea.Msg {
	// Check connection on slow tick (1Hz)
	if !m.redisClient.CheckConnection() {
		return dataUpdateMsg{err: fmt.Errorf("redis: %s", m.redisClient.LastError)}
	}

	vehicle, _ := m.redisClient.GetVehicleData()
	battery0, _ := m.redisClient.GetBatteryData(0)
	battery1, _ := m.redisClient.GetBatteryData(1)
	navigation, _ := m.redisClient.GetNavigationData()
	dashboard, _ := m.redisClient.GetDashboardData()
	internet, _ := m.redisClient.GetInternetData()
	bluetooth, _ := m.redisClient.GetBluetoothData()
	speedLimit, _ := m.redisClient.GetSpeedLimitData()
	settings, _ := m.redisClient.GetSettingsData()
	ota, _ := m.redisClient.GetOtaData()

	return dataUpdateMsg{
		vehicle:    vehicle,
		battery0:   battery0,
		battery1:   battery1,
		navigation: navigation,
		dashboard:  dashboard,
		internet:   internet,
		bluetooth:  bluetooth,
		speedLimit: speedLimit,
		settings:   settings,
		ota:        ota,
	}
}

func fetchAllDataCmd() tea.Cmd {
	return func() tea.Msg {
		return fetchDataMsg{}
	}
}

func fetchSlowDataCmd() tea.Cmd {
	return func() tea.Msg {
		return fetchSlowDataMsg{}
	}
}

func (m Model) calculateRoute(startLat, startLon, endLat, endLon float64) tea.Cmd {
	return func() tea.Msg {
		route, err := m.valhallaClient.CalculateRoute(startLat, startLon, endLat, endLon)
		return routeCalculatedMsg{
			route: route,
			err:   err,
		}
	}
}

// buildSettingsMenu constructs the settings menu tree with current values.
func buildSettingsMenu(s *redis.SettingsData) components.MenuState {
	alarmDur := "10"
	switch s.AlarmDuration {
	case 20:
		alarmDur = "20"
	case 30:
		alarmDur = "30"
	default:
		alarmDur = "10"
	}

	items := []components.MenuItem{
		{
			Label: "Display",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "Theme", Key: "dashboard.theme", Type: components.MenuCycle,
					Options: []string{"dark", "light", "auto"}, Value: s.Theme},
				{Label: "Language", Key: "dashboard.language", Type: components.MenuCycle,
					Options: []string{"en", "de"}, Value: s.Language},
				{Label: "Default Screen", Key: "dashboard.mode", Type: components.MenuCycle,
					Options: []string{"speedometer", "navigation"}, Value: s.Mode},
				{Label: "Raw Speed", Key: "dashboard.show-raw-speed", Type: components.MenuToggle,
					Value: boolStr(s.ShowRawSpeed)},
				{Label: "Power Display", Key: "dashboard.power-display-mode", Type: components.MenuCycle,
					Options: []string{"kw", "amps"}, Value: s.PowerDisplayMode},
			},
		},
		{
			Label: "Status Bar",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "GPS", Key: "dashboard.show-gps", Type: components.MenuCycle,
					Options: []string{"always", "active-or-error", "error", "never"}, Value: s.ShowGps},
				{Label: "Bluetooth", Key: "dashboard.show-bluetooth", Type: components.MenuCycle,
					Options: []string{"always", "active-or-error", "error", "never"}, Value: s.ShowBluetooth},
				{Label: "Cloud", Key: "dashboard.show-cloud", Type: components.MenuCycle,
					Options: []string{"always", "active-or-error", "error", "never"}, Value: s.ShowCloud},
				{Label: "Internet", Key: "dashboard.show-internet", Type: components.MenuCycle,
					Options: []string{"always", "active-or-error", "error", "never"}, Value: s.ShowInternet},
				{Label: "Clock", Key: "dashboard.show-clock", Type: components.MenuCycle,
					Options: []string{"always", "never"}, Value: s.ShowClock},
			},
		},
		{
			Label: "Battery",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "Dual Battery", Key: "scooter.dual-battery", Type: components.MenuToggle,
					Value: boolStr(s.DualBattery)},
				{Label: "Display", Key: "dashboard.battery-display-mode", Type: components.MenuCycle,
					Options: []string{"percentage", "range"}, Value: s.BatteryDisplayMode},
			},
		},
		{
			Label: "Blinker Style",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "Style", Key: "dashboard.blinker-style", Type: components.MenuCycle,
					Options: []string{"icon", "overlay"}, Value: s.BlinkerStyle},
			},
		},
		{
			Label: "Alarm",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "Enabled", Key: "alarm.enabled", Type: components.MenuToggle,
					Value: boolStr(s.AlarmEnabled)},
				{Label: "Honk", Key: "alarm.honk", Type: components.MenuToggle,
					Value: boolStr(s.AlarmHonk)},
				{Label: "Duration", Key: "alarm.duration", Type: components.MenuCycle,
					Options: []string{"10", "20", "30"}, Value: alarmDur},
			},
		},
		{
			Label: "Vehicle",
			Type:  components.MenuSubmenu,
			Children: []components.MenuItem{
				{Label: "Horn", Key: "scooter.enable-horn", Type: components.MenuCycle,
					Options: []string{"true", "false", "in-drive"}, Value: s.EnableHorn},
			},
		},
		{Label: "Reset Trip", Type: components.MenuAction},
		{Label: "About & Licenses", Type: components.MenuAction},
	}

	return components.NewMenuState("Settings", items)
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// updateMenuValues updates the displayed values in the menu tree
// without resetting cursor/scroll position.
func updateMenuValues(state *components.MenuState, s *redis.SettingsData) {
	vals := map[string]string{
		"dashboard.theme":                s.Theme,
		"dashboard.language":             s.Language,
		"dashboard.mode":                 s.Mode,
		"dashboard.show-raw-speed":       boolStr(s.ShowRawSpeed),
		"dashboard.show-gps":             s.ShowGps,
		"dashboard.show-bluetooth":       s.ShowBluetooth,
		"dashboard.show-cloud":           s.ShowCloud,
		"dashboard.show-internet":        s.ShowInternet,
		"dashboard.show-clock":           s.ShowClock,
		"dashboard.blinker-style":        s.BlinkerStyle,
		"dashboard.battery-display-mode": s.BatteryDisplayMode,
		"dashboard.power-display-mode":   s.PowerDisplayMode,
		"scooter.dual-battery":           boolStr(s.DualBattery),
		"scooter.enable-horn":            s.EnableHorn,
		"alarm.enabled":                  boolStr(s.AlarmEnabled),
		"alarm.honk":                     boolStr(s.AlarmHonk),
		"alarm.duration":                 fmt.Sprintf("%d", s.AlarmDuration),
	}

	for i := range state.Stack {
		updateItemValues(state.Stack[i].Items, vals)
	}
}

func updateItemValues(items []components.MenuItem, vals map[string]string) {
	for i := range items {
		if items[i].Key != "" {
			if v, ok := vals[items[i].Key]; ok {
				items[i].Value = v
			}
		}
		if len(items[i].Children) > 0 {
			updateItemValues(items[i].Children, vals)
		}
	}
}
