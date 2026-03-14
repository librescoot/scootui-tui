package redis

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// Client wraps Redis client with helper methods and reconnection.
type Client struct {
	client    *redis.Client
	ctx       context.Context
	pubsub    *redis.PubSub
	addr      string
	Connected bool
	LastError string
}

// NewClient creates a Redis client. Does not fail — go-redis
// reconnects automatically, and Connected is updated on each fetch.
func NewClient(addr string) *Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		MaxRetries:   3,
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx := context.Background()

	c := &Client{
		client: rdb,
		ctx:    ctx,
		addr:   addr,
	}

	// Test initial connection (non-fatal)
	if err := rdb.Ping(ctx).Err(); err != nil {
		c.Connected = false
		c.LastError = err.Error()
	} else {
		c.Connected = true
	}

	return c
}

// CheckConnection pings Redis and updates Connected state.
func (c *Client) CheckConnection() bool {
	if err := c.client.Ping(c.ctx).Err(); err != nil {
		c.Connected = false
		c.LastError = err.Error()
		return false
	}
	c.Connected = true
	c.LastError = ""
	return true
}

// Close closes the Redis connection
func (c *Client) Close() error {
	if c.pubsub != nil {
		c.pubsub.Close()
	}
	return c.client.Close()
}

// Ping checks if Redis is connected
func (c *Client) Ping() error {
	return c.client.Ping(c.ctx).Err()
}

// Subscribe subscribes to Redis PUBSUB channels
func (c *Client) Subscribe(channels ...string) error {
	c.pubsub = c.client.Subscribe(c.ctx, channels...)
	return c.pubsub.Ping(c.ctx)
}

// ReceiveMessage receives next PUBSUB message (blocking)
func (c *Client) ReceiveMessage() (string, string, error) {
	if c.pubsub == nil {
		return "", "", fmt.Errorf("not subscribed to any channels")
	}

	msg, err := c.pubsub.ReceiveMessage(c.ctx)
	if err != nil {
		return "", "", err
	}

	return msg.Channel, msg.Payload, nil
}

// Subscription wraps a Redis PUBSUB subscription for a single channel.
type Subscription struct {
	pubsub *redis.PubSub
	ctx    context.Context
}

// SubscribeChannel creates a new subscription on a single channel.
// Returns nil if subscription fails.
func (c *Client) SubscribeChannel(channel string) *Subscription {
	sub := c.client.Subscribe(c.ctx, channel)
	if err := sub.Ping(c.ctx); err != nil {
		sub.Close()
		return nil
	}
	return &Subscription{pubsub: sub, ctx: c.ctx}
}

// ReceiveMessage blocks until the next message and returns the payload.
// Returns empty string on error.
func (s *Subscription) ReceiveMessage() string {
	msg, err := s.pubsub.ReceiveMessage(s.ctx)
	if err != nil {
		return ""
	}
	return msg.Payload
}

// HGet gets a hash field value
func (c *Client) HGet(key, field string) (string, error) {
	return c.client.HGet(c.ctx, key, field).Result()
}

// HGetAll gets all hash fields
func (c *Client) HGetAll(key string) (map[string]string, error) {
	return c.client.HGetAll(c.ctx, key).Result()
}

// SMembers gets all members of a set
func (c *Client) SMembers(key string) ([]string, error) {
	return c.client.SMembers(c.ctx, key).Result()
}

// Helper functions to parse Redis values

// ParseInt parses a string to int64
func ParseInt(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

// ParseFloat parses a string to float64
func ParseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
}

// ParseBool parses "true"/"false" or "on"/"off" to bool
func ParseBool(s string) bool {
	s = strings.ToLower(s)
	return s == "true" || s == "on" || s == "1"
}

// ParseTime parses ISO8601 timestamp
func ParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		// Try alternative format
		t, _ = time.Parse("2006-01-02T15:04:05.999999", s)
	}
	return t
}

// GetVehicleData fetches vehicle data from Redis
func (c *Client) GetVehicleData() (*VehicleData, error) {
	data, err := c.HGetAll("vehicle")
	if err != nil {
		return nil, err
	}

	return &VehicleData{
		BlinkerState:         BlinkerState(data["blinker:state"]),
		BlinkerSwitch:        BlinkerState(data["blinker:switch"]),
		BrakeLeft:            OnOffState(data["brake:left"]),
		BrakeRight:           OnOffState(data["brake:right"]),
		Kickstand:            UpDownState(data["kickstand"]),
		State:                VehicleState(data["state"]),
		HandleBarLockSensor:  OnOffState(data["handlebar:lock-sensor"]),
		HandleBarPosition:    data["handlebar:position"],
		SeatboxButton:        OnOffState(data["seatbox:button"]),
		SeatboxLock:          OpenClosedState(data["seatbox:lock"]),
		HornButton:           OnOffState(data["horn-button"]),
		IsUnableToDrive:      OnOffState(data["unable-to-drive"]),
		AutoStandbyRemaining: int(ParseInt(data["auto-standby-remaining"])),
	}, nil
}

// GetEngineData fetches engine data from Redis
func (c *Client) GetEngineData() (*EngineData, error) {
	data, err := c.HGetAll("engine-ecu")
	if err != nil {
		return nil, err
	}

	return &EngineData{
		PowerState:      OnOffState(data["state"]),
		Kers:            OnOffState(data["kers"]),
		KersReasonOff:   data["kers-reason-off"],
		MotorVoltage:    ParseInt(data["motor:voltage"]),
		MotorCurrent:    ParseInt(data["motor:current"]),
		RPM:             ParseInt(data["rpm"]),
		Speed:           ParseInt(data["speed"]),
		RawSpeed:        ParseInt(data["raw-speed"]),
		Throttle:        OnOffState(data["throttle"]),
		FirmwareVersion: data["fw-version"],
		Odometer:        ParseFloat(data["odometer"]),
		Temperature:     ParseFloat(data["temperature"]),
	}, nil
}

// GetBatteryData fetches battery data from Redis
func (c *Client) GetBatteryData(id int) (*BatteryData, error) {
	key := fmt.Sprintf("battery:%d", id)
	data, err := c.HGetAll(key)
	if err != nil {
		return nil, err
	}

	// Get fault set
	faultKey := fmt.Sprintf("battery:%d:fault", id)
	faultStrs, _ := c.SMembers(faultKey)
	faults := make([]int, 0, len(faultStrs))
	for _, f := range faultStrs {
		if fault, err := strconv.Atoi(f); err == nil {
			faults = append(faults, fault)
		}
	}

	return &BatteryData{
		ID:                id,
		Present:           ParseBool(data["present"]),
		State:             BatteryState(data["state"]),
		Voltage:           ParseInt(data["voltage"]),
		Current:           ParseInt(data["current"]),
		Charge:            int(ParseInt(data["charge"])),
		Temperature0:      int(ParseInt(data["temperature:0"])),
		Temperature1:      int(ParseInt(data["temperature:1"])),
		Temperature2:      int(ParseInt(data["temperature:2"])),
		Temperature3:      int(ParseInt(data["temperature:3"])),
		TemperatureState:  data["temperature-state"],
		CycleCount:        int(ParseInt(data["cycle-count"])),
		StateOfHealth:     int(ParseInt(data["state-of-health"])),
		SerialNumber:      data["serial-number"],
		ManufacturingDate: data["manufacturing-date"],
		FirmwareVersion:   data["fw-version"],
		Faults:            faults,
	}, nil
}

// GetGpsData fetches GPS data from Redis
func (c *Client) GetGpsData() (*GpsData, error) {
	data, err := c.HGetAll("gps")
	if err != nil {
		return nil, err
	}

	// Parse timestamp (prefer "updated" over "timestamp")
	timeStr := data["updated"]
	if timeStr == "" {
		timeStr = data["timestamp"]
	}

	return &GpsData{
		Latitude:  ParseFloat(data["latitude"]),
		Longitude: ParseFloat(data["longitude"]),
		Course:    ParseFloat(data["course"]),
		Speed:     ParseFloat(data["speed"]),
		Altitude:  ParseFloat(data["altitude"]),
		Updated:   ParseTime(timeStr),
		State:     GpsState(data["state"]),
	}, nil
}

// GetNavigationData fetches navigation destination from Redis
func (c *Client) GetNavigationData() (*NavigationData, error) {
	data, err := c.HGetAll("navigation")
	if err != nil {
		return nil, err
	}

	return &NavigationData{
		Latitude:  data["latitude"],
		Longitude: data["longitude"],
		Address:   data["address"],
		Timestamp: ParseTime(data["timestamp"]),
	}, nil
}

// GetDashboardData fetches dashboard data from Redis
func (c *Client) GetDashboardData() (*DashboardData, error) {
	data, err := c.HGetAll("dashboard")
	if err != nil {
		return nil, err
	}

	return &DashboardData{
		Brightness: ParseFloat(data["brightness"]),
		Backlight:  int(ParseInt(data["backlight"])),
		Theme:      data["theme"],
		Mode:       data["mode"],
		Debug:      data["debug"],
		Time:       time.Now(), // Use current time
	}, nil
}

// GetInternetData fetches internet/modem data from Redis
func (c *Client) GetInternetData() (*InternetData, error) {
	data, err := c.HGetAll("internet")
	if err != nil {
		return nil, err
	}

	return &InternetData{
		ModemState:    ModemState(data["modem-state"]),
		UnuCloud:      ConnectedState(data["unu-cloud"]),
		Status:        ConnectedState(data["status"]),
		IPAddress:     data["ip-address"],
		AccessTech:    data["access-tech"],
		SignalQuality: int(ParseInt(data["signal-quality"])),
		SimIMEI:       data["sim-imei"],
		SimIMSI:       data["sim-imsi"],
		SimICCID:      data["sim-iccid"],
	}, nil
}

// GetBluetoothData fetches Bluetooth data from Redis
func (c *Client) GetBluetoothData() (*BluetoothData, error) {
	data, err := c.HGetAll("ble")
	if err != nil {
		return nil, err
	}

	return &BluetoothData{
		Status:        ConnectedState(data["status"]),
		MacAddress:    data["mac-address"],
		PinCode:       data["pin-code"],
		ServiceHealth: data["service-health"],
		ServiceError:  data["service-error"],
		LastUpdate:    ParseTime(data["last-update"]),
	}, nil
}

// HSet sets a hash field value
func (c *Client) HSet(key, field, value string) error {
	return c.client.HSet(c.ctx, key, field, value).Err()
}

// GetSettingsData fetches settings from Redis
func (c *Client) GetSettingsData() (*SettingsData, error) {
	data, err := c.HGetAll("settings")
	if err != nil {
		return nil, err
	}

	s := DefaultSettings()
	if v := data["dashboard.theme"]; v != "" {
		s.Theme = v
	}
	if v := data["dashboard.language"]; v != "" {
		s.Language = v
	}
	if v := data["dashboard.show-gps"]; v != "" {
		s.ShowGps = v
	}
	if v := data["dashboard.show-bluetooth"]; v != "" {
		s.ShowBluetooth = v
	}
	if v := data["dashboard.show-cloud"]; v != "" {
		s.ShowCloud = v
	}
	if v := data["dashboard.show-internet"]; v != "" {
		s.ShowInternet = v
	}
	if v := data["dashboard.show-clock"]; v != "" {
		s.ShowClock = v
	}
	if v := data["scooter.dual-battery"]; v != "" {
		s.DualBattery = ParseBool(v)
	}
	if v := data["dashboard.blinker-style"]; v != "" {
		s.BlinkerStyle = v
	}
	if v := data["alarm.enabled"]; v != "" {
		s.AlarmEnabled = ParseBool(v)
	}
	if v := data["alarm.honk"]; v != "" {
		s.AlarmHonk = ParseBool(v)
	}
	if v := data["alarm.duration"]; v != "" {
		s.AlarmDuration = int(ParseInt(v))
	}
	if v := data["dashboard.valhalla-url"]; v != "" {
		s.ValhallaURL = v
	}
	if v := data["dashboard.mode"]; v != "" {
		s.Mode = v
	}
	if v := data["dashboard.show-raw-speed"]; v != "" {
		s.ShowRawSpeed = ParseBool(v)
	}
	if v := data["dashboard.battery-display-mode"]; v != "" {
		s.BatteryDisplayMode = v
	}
	if v := data["dashboard.power-display-mode"]; v != "" {
		s.PowerDisplayMode = v
	}
	if v := data["scooter.auto-standby-seconds"]; v != "" {
		s.AutoStandbySeconds = int(ParseInt(v))
	}
	if v := data["scooter.enable-horn"]; v != "" {
		s.EnableHorn = v
	}

	return s, nil
}

// SetSetting writes a single setting to the settings hash.
func (c *Client) SetSetting(key, value string) error {
	return c.HSet("settings", key, value)
}

// GetOtaData fetches OTA update state from Redis
func (c *Client) GetOtaData() (*OtaData, error) {
	data, err := c.HGetAll("ota")
	if err != nil {
		return nil, err
	}

	return &OtaData{
		DbcStatus:           data["dbc-status"],
		DbcUpdateVersion:    data["dbc-update-version"],
		DbcDownloadProgress: int(ParseInt(data["dbc-download-progress"])),
		DbcInstallProgress:  int(ParseInt(data["dbc-install-progress"])),
		DbcError:            data["dbc-error"],
		MdbStatus:           data["mdb-status"],
		MdbUpdateVersion:    data["mdb-update-version"],
		MdbDownloadProgress: int(ParseInt(data["mdb-download-progress"])),
		MdbInstallProgress:  int(ParseInt(data["mdb-install-progress"])),
		MdbError:            data["mdb-error"],
	}, nil
}

// GetSpeedLimitData fetches speed limit data from Redis
func (c *Client) GetSpeedLimitData() (*SpeedLimitData, error) {
	data, err := c.HGetAll("speed-limit")
	if err != nil {
		return nil, err
	}

	return &SpeedLimitData{
		SpeedLimit: data["speed-limit"],
		RoadName:   data["road-name"],
		RoadType:   data["road-type"],
	}, nil
}
