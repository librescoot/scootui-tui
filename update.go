package main

import (
	"scootui-tui/components"
	"scootui-tui/input"
	"scootui-tui/redis"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		if !m.fixedSize {
			m.width = msg.Width
			m.height = msg.Height
		}
		return m, nil

	case tickMsg:
		m.toasts.Prune()
		if m.vehicle != nil && m.engine != nil && m.gps != nil {
			m.updateTrip()
		}
		var cmds []tea.Cmd
		cmds = append(cmds, fastTickCmd(), tea.Cmd(func() tea.Msg {
			return fetchFastData(m)
		}))
		// Check gesture detector for holds and pending taps
		for _, evt := range m.gestures.CheckHolds() {
			e := evt
			cmds = append(cmds, func() tea.Msg { return gestureMsg(e) })
		}
		for _, evt := range m.gestures.FlushPending() {
			e := evt
			cmds = append(cmds, func() tea.Msg { return gestureMsg(e) })
		}
		return m, tea.Batch(cmds...)

	case slowTickMsg:
		return m, tea.Batch(slowTickCmd(), tea.Cmd(func() tea.Msg {
			return fetchSlowData(m)
		}))

	case fetchDataMsg:
		return m, tea.Cmd(func() tea.Msg {
			return fetchSlowData(m)
		})

	case fetchSlowDataMsg:
		return m, tea.Cmd(func() tea.Msg {
			return fetchSlowData(m)
		})

	case blinkerTickMsg:
		m.blinkerFlash = !m.blinkerFlash
		return m, blinkerTickCmd()

	case buttonEventMsg:
		return m.handleButtonEvent(msg)

	case buttonRetryMsg:
		return m, m.listenButtons()

	case gestureMsg:
		return m.handleGesture(input.GestureEvent(msg))

	case dataUpdateMsg:
		return m.handleDataUpdate(msg)

	case routeCalculatedMsg:
		if msg.err != nil {
			m.routeError = msg.err.Error()
			m.route = nil
			m.toasts.Show("Route error: "+msg.err.Error(), components.ToastError)
		} else {
			m.route = msg.route
			m.routeError = ""
			m.toasts.Show("Route calculated", components.ToastSuccess)
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handleButtonEvent(msg buttonEventMsg) (tea.Model, tea.Cmd) {
	var gestureCmds []tea.Cmd

	if msg.pressed {
		events := m.gestures.Press(msg.control)
		for _, evt := range events {
			e := evt
			gestureCmds = append(gestureCmds, func() tea.Msg { return gestureMsg(e) })
		}
	} else {
		events := m.gestures.Release(msg.control)
		for _, evt := range events {
			e := evt
			gestureCmds = append(gestureCmds, func() tea.Msg { return gestureMsg(e) })
		}
	}

	// Always re-listen for next button event
	gestureCmds = append(gestureCmds, m.listenButtons())
	return m, tea.Batch(gestureCmds...)
}

// handleGesture maps physical gestures to UI actions.
//
// Interaction model (matching Flutter scootui):
//
// Main screens (Cluster / Navigation):
//
//	Left brake tap:         scroll through screens (cluster <-> nav)
//	Left brake double-tap:  open settings menu
//	Right brake tap:        (unused on main screens)
//	Seatbox tap:            (reserved)
//	Seatbox double-press:   toggle hazard lights (future)
//	Seatbox hold:           show shortcut menu, items auto-cycle
//	Seatbox release:        confirm shortcut menu selection
//
// Settings screen:
//
//	Left brake tap:         scroll down
//	Left brake hold:        fast scroll down (repeated via tick)
//	Right brake tap:        select / confirm
//	Right brake hold:       go back one level (or exit settings)
//
// About screen:
//
//	Left brake tap:         scroll down
//	Left brake hold:        fast scroll
//	Right brake tap:        exit to cluster
func (m Model) handleGesture(evt input.GestureEvent) (tea.Model, tea.Cmd) {
	switch evt.Control {
	case input.ControlLeft:
		return m.handleLeftGesture(evt.Gesture)
	case input.ControlRight:
		return m.handleRightGesture(evt.Gesture)
	case input.ControlSeat:
		return m.handleSeatGesture(evt.Gesture)
	}
	return m, nil
}

func (m Model) handleLeftGesture(g input.Gesture) (tea.Model, tea.Cmd) {
	switch g {
	case input.GestureTap:
		switch m.activeScreen {
		case ScreenCluster:
			m.activeScreen = ScreenNavigation
		case ScreenNavigation:
			m.activeScreen = ScreenCluster
		case ScreenSettings:
			m.menuState.MoveDown()
		case ScreenAbout:
			m.aboutScroll++
		}

	case input.GestureDoubleTap:
		// Open settings only when parked (matching Flutter scootui)
		if m.activeScreen == ScreenCluster || m.activeScreen == ScreenNavigation {
			if m.vehicle.State == "parked" || m.vehicle.State == "stand-by" {
				m.activeScreen = ScreenSettings
			}
		}

	case input.GestureHold:
		switch m.activeScreen {
		case ScreenSettings:
			m.menuState.MoveDown()
		case ScreenAbout:
			m.aboutScroll++
		}
	}

	return m, nil
}

func (m Model) handleRightGesture(g input.Gesture) (tea.Model, tea.Cmd) {
	switch g {
	case input.GestureTap:
		switch m.activeScreen {
		case ScreenSettings:
			item := m.menuState.Enter()
			if item != nil {
				return m.handleMenuAction(item)
			}
		case ScreenAbout:
			m.activeScreen = ScreenCluster
		}

	case input.GestureHold:
		// Back / escape
		switch m.activeScreen {
		case ScreenSettings:
			if !m.menuState.Back() {
				m.activeScreen = ScreenCluster
			}
		case ScreenAbout:
			m.activeScreen = ScreenCluster
		case ScreenNavigation:
			m.activeScreen = ScreenCluster
		}
	}

	return m, nil
}

func (m Model) handleSeatGesture(g input.Gesture) (tea.Model, tea.Cmd) {
	switch g {
	case input.GestureTap:
		// Cycle between main screens only (Cluster <-> Nav)
		if m.activeScreen < numMainScreens {
			m.activeScreen = (m.activeScreen + 1) % numMainScreens
		}

	case input.GestureDoubleTap:
		// Toggle hazard lights (future: send to Redis)
	}

	return m, nil
}

// Keyboard handling (for development/testing)

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "q", "ctrl+c":
		m.quitting = true
		return m, tea.Quit
	case "tab":
		if m.activeScreen < numMainScreens {
			m.activeScreen = (m.activeScreen + 1) % numMainScreens
		}
		return m, nil
	case "1":
		m.activeScreen = ScreenCluster
		return m, nil
	case "2":
		m.activeScreen = ScreenNavigation
		return m, nil
	case "3":
		// Settings only when parked (keyboard override for dev)
		m.activeScreen = ScreenSettings
		return m, nil
	case "4":
		m.activeScreen = ScreenAbout
		m.aboutScroll = 0
		return m, nil
	}

	switch m.activeScreen {
	case ScreenCluster:
		return m.handleClusterKey(key)
	case ScreenNavigation:
		return m.handleNavigationKey(key)
	case ScreenSettings:
		return m.handleSettingsKey(key)
	case ScreenAbout:
		return m.handleAboutKey(key)
	}

	return m, nil
}

func (m Model) handleClusterKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "d":
		m.debugMode = !m.debugMode
	case "r":
		return m, fetchAllDataCmd()
	}
	return m, nil
}

func (m Model) handleNavigationKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.activeScreen = ScreenCluster
	case "r":
		return m, fetchAllDataCmd()
	}
	return m, nil
}

func (m Model) handleSettingsKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		if !m.menuState.Back() {
			m.activeScreen = ScreenCluster
		}
	case "j", "down":
		m.menuState.MoveDown()
	case "k", "up":
		m.menuState.MoveUp()
	case "enter", " ":
		item := m.menuState.Enter()
		if item != nil {
			return m.handleMenuAction(item)
		}
	}
	return m, nil
}

func (m Model) handleAboutKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc":
		m.activeScreen = ScreenCluster
	case "j", "down":
		m.aboutScroll++
	case "k", "up":
		if m.aboutScroll > 0 {
			m.aboutScroll--
		}
	}
	return m, nil
}

func (m Model) handleMenuAction(item *components.MenuItem) (tea.Model, tea.Cmd) {
	switch item.Type {
	case components.MenuToggle:
		newVal := "true"
		if item.Value == "true" || item.Value == "on" {
			newVal = "false"
		}
		item.Value = newVal
		if m.redisClient != nil && item.Key != "" {
			m.redisClient.SetSetting(item.Key, newVal)
		}

	case components.MenuCycle:
		if len(item.Options) > 0 {
			current := 0
			for i, opt := range item.Options {
				if opt == item.Value {
					current = i
					break
				}
			}
			next := (current + 1) % len(item.Options)
			item.Value = item.Options[next]
			if m.redisClient != nil && item.Key != "" {
				m.redisClient.SetSetting(item.Key, item.Value)
			}
		}

	case components.MenuAction:
		switch item.Label {
		case "Reset Trip":
			m.trip = &redis.TripData{}
		case "About & Licenses":
			m.activeScreen = ScreenAbout
			m.aboutScroll = 0
		}
	}

	return m, nil
}

func (m Model) handleDataUpdate(msg dataUpdateMsg) (tea.Model, tea.Cmd) {
	m.toasts.Prune()

	if msg.err != nil {
		if m.redisClient.Connected {
			// Was connected, now failing — show error toast
			m.toasts.Show("Redis connection lost", components.ToastError)
		}
		m.redisClient.Connected = false
		m.redisClient.LastError = msg.err.Error()
		return m, nil
	}

	if !m.redisClient.Connected {
		// Was disconnected, now recovered
		m.toasts.Show("Redis connected", components.ToastSuccess)
	}
	m.redisClient.Connected = true
	m.redisClient.LastError = ""

	var needsRoute bool
	if msg.navigation != nil && msg.navigation.HasDestination() {
		if m.navigation == nil || !m.navigation.HasDestination() ||
			m.navigation.Latitude != msg.navigation.Latitude ||
			m.navigation.Longitude != msg.navigation.Longitude {
			needsRoute = true
			addr := msg.navigation.Address
			if addr == "" {
				addr = "new destination"
			}
			m.toasts.Show("Navigating to "+addr, components.ToastInfo)
		}
	} else if msg.navigation != nil && !msg.navigation.HasDestination() {
		if m.route != nil {
			m.toasts.Show("Navigation cleared", components.ToastInfo)
		}
		m.route = nil
	}

	if msg.vehicle != nil {
		m.vehicle = msg.vehicle
	}
	if msg.engine != nil {
		m.engine = msg.engine
	}
	if msg.battery0 != nil {
		m.battery0 = msg.battery0
	}
	if msg.battery1 != nil {
		m.battery1 = msg.battery1
	}
	if msg.gps != nil {
		m.gps = msg.gps
		if m.route != nil && m.gps.HasRecentFix() {
			m.route.UpdateRouteProgress(m.gps.Latitude, m.gps.Longitude)
		}
	}
	if msg.navigation != nil {
		m.navigation = msg.navigation
	}
	if msg.dashboard != nil {
		m.dashboard = msg.dashboard
	}
	if msg.internet != nil {
		m.internet = msg.internet
	}
	if msg.bluetooth != nil {
		m.bluetooth = msg.bluetooth
	}
	if msg.speedLimit != nil {
		m.speedLimit = msg.speedLimit
	}
	if msg.settings != nil {
		m.settings = msg.settings
		updateMenuValues(&m.menuState, m.settings)
		if m.settings.ValhallaURL != "" {
			m.valhallaClient.SetEndpoint(m.settings.ValhallaURL)
		}
	}
	if msg.ota != nil {
		m.ota = msg.ota
	}

	// Calculate route if destination changed, or if we have a destination
	// but no route yet (e.g. startup, or GPS fix arrived after destination).
	// Don't retry if we already have a recent error (avoid hammering the API).
	if !needsRoute && m.route == nil && m.routeError == "" && m.navigation.HasDestination() {
		needsRoute = true
	}

	if needsRoute && m.gps.HasRecentFix() {
		endLat, _ := m.navigation.LatitudeFloat()
		endLon, _ := m.navigation.LongitudeFloat()
		return m, m.calculateRoute(m.gps.Latitude, m.gps.Longitude, endLat, endLon)
	}

	return m, nil
}

func (m *Model) updateTrip() {
	if m.vehicle.State == "ready-to-drive" && m.trip.StartTime.IsZero() {
		m.trip.StartTime = time.Now()
		m.trip.Distance = 0
		m.trip.TotalSpeed = 0
		m.trip.SpeedSamples = 0
	}

	if m.vehicle.State != "ready-to-drive" && !m.trip.StartTime.IsZero() {
		m.trip = &redis.TripData{}
	}

	if !m.trip.StartTime.IsZero() && m.gps.HasRecentFix() {
		speed := m.engine.SpeedKmh()
		if speed > 0 {
			m.trip.TotalSpeed += speed
			m.trip.SpeedSamples++
			m.trip.Distance += speed * 1000.0 / 3600.0
		}
	}
}
