package redis

import (
	"strconv"
	"time"
)

type VehicleState string

const (
	VehicleStandBy               VehicleState = "stand-by"
	VehicleReadyToDrive          VehicleState = "ready-to-drive"
	VehicleOff                   VehicleState = "off"
	VehicleParked                VehicleState = "parked"
	VehicleBooting               VehicleState = "booting"
	VehicleShuttingDown          VehicleState = "shutting-down"
	VehicleHibernating           VehicleState = "hibernating"
	VehicleHibernatingImminent   VehicleState = "hibernating-imminent"
	VehicleSuspending            VehicleState = "suspending"
	VehicleSuspendingImminent    VehicleState = "suspending-imminent"
	VehicleUpdating              VehicleState = "updating"
	VehicleWaitingHibernation    VehicleState = "waiting-hibernation"
	VehicleWaitingHibernationAdv VehicleState = "waiting-hibernation-advanced"
	VehicleWaitingHibernationSB  VehicleState = "waiting-hibernation-seatbox"
	VehicleWaitingHibernationCnf VehicleState = "waiting-hibernation-confirm"
)

type BlinkerState string

const (
	BlinkerOff   BlinkerState = "off"
	BlinkerLeft  BlinkerState = "left"
	BlinkerRight BlinkerState = "right"
	BlinkerBoth  BlinkerState = "both"
)

type OnOffState string

const (
	StateOn  OnOffState = "on"
	StateOff OnOffState = "off"
)

type UpDownState string

const (
	StateUp   UpDownState = "up"
	StateDown UpDownState = "down"
)

type OpenClosedState string

const (
	StateOpen   OpenClosedState = "open"
	StateClosed OpenClosedState = "closed"
)

type BatteryState string

const (
	BatteryUnknown BatteryState = "unknown"
	BatteryAsleep  BatteryState = "asleep"
	BatteryIdle    BatteryState = "idle"
	BatteryActive  BatteryState = "active"
)

type GpsState string

const (
	GpsOff            GpsState = "off"
	GpsSearching      GpsState = "searching"
	GpsFixEstablished GpsState = "fix-established"
	GpsError          GpsState = "error"
)

type ModemState string

const (
	ModemOff          ModemState = "off"
	ModemDisconnected ModemState = "disconnected"
	ModemConnected    ModemState = "connected"
)

type ConnectedState string

const (
	Connected    ConnectedState = "connected"
	Disconnected ConnectedState = "disconnected"
)

type VehicleData struct {
	BlinkerState         BlinkerState
	BlinkerSwitch        BlinkerState
	BrakeLeft            OnOffState
	BrakeRight           OnOffState
	Kickstand            UpDownState
	State                VehicleState
	HandleBarLockSensor  OnOffState
	HandleBarPosition    string
	SeatboxButton        OnOffState
	SeatboxLock          OpenClosedState
	HornButton           OnOffState
	IsUnableToDrive      OnOffState
	AutoStandbyRemaining int
}

type EngineData struct {
	PowerState      OnOffState
	Kers            OnOffState
	KersReasonOff   string
	MotorVoltage    int64 // millivolts
	MotorCurrent    int64 // milliamps
	RPM             int64
	Speed           int64 // km/h * 100
	RawSpeed        int64 // km/h * 100
	Throttle        OnOffState
	FirmwareVersion string
	Odometer        float64 // meters
	Temperature     float64 // celsius
}

func (e *EngineData) PowerOutput() float64 {
	return float64(e.MotorVoltage) * float64(e.MotorCurrent) / 1000000.0
}

func (e *EngineData) SpeedKmh() float64 {
	return float64(e.Speed)
}

type BatteryData struct {
	ID                int
	Present           bool
	State             BatteryState
	Voltage           int64 // millivolts
	Current           int64 // milliamps
	Charge            int   // 0-100%
	Temperature0      int   // celsius
	Temperature1      int   // celsius
	Temperature2      int   // celsius
	Temperature3      int   // celsius
	TemperatureState  string
	CycleCount        int
	StateOfHealth     int // 0-100%
	SerialNumber      string
	ManufacturingDate string
	FirmwareVersion   string
	Faults            []int
}

type GpsData struct {
	Latitude  float64
	Longitude float64
	Course    float64 // degrees
	Speed     float64 // km/h
	Altitude  float64 // meters
	Updated   time.Time
	State     GpsState
}

func (g *GpsData) HasRecentFix() bool {
	return g.State == GpsFixEstablished && time.Since(g.Updated) < 10*time.Second
}

type NavigationData struct {
	Latitude  string
	Longitude string
	Address   string
	Timestamp time.Time
}

func (n *NavigationData) HasDestination() bool {
	return n.Latitude != "" && n.Longitude != ""
}

func (n *NavigationData) LatitudeFloat() (float64, error) {
	return strconv.ParseFloat(n.Latitude, 64)
}

func (n *NavigationData) LongitudeFloat() (float64, error) {
	return strconv.ParseFloat(n.Longitude, 64)
}

type DashboardData struct {
	Brightness float64
	Backlight  int
	Theme      string
	Mode       string
	Debug      string
	Time       time.Time
}

type InternetData struct {
	ModemState    ModemState
	UnuCloud      ConnectedState
	Status        ConnectedState
	IPAddress     string
	AccessTech    string // "2G", "3G", "4G", "5G"
	SignalQuality int    // 0-100
	SimIMEI       string
	SimIMSI       string
	SimICCID      string
}

func (i *InternetData) SignalBars() int {
	if i.SignalQuality >= 80 {
		return 4
	} else if i.SignalQuality >= 60 {
		return 3
	} else if i.SignalQuality >= 40 {
		return 2
	} else if i.SignalQuality >= 20 {
		return 1
	}
	return 0
}

type BluetoothData struct {
	Status        ConnectedState
	MacAddress    string
	PinCode       string
	ServiceHealth string
	ServiceError  string
	LastUpdate    time.Time
}

type SpeedLimitData struct {
	SpeedLimit string
	RoadName   string
	RoadType   string
}

func (s *SpeedLimitData) HasSpeedLimit() bool {
	return s.SpeedLimit != "" && s.SpeedLimit != "none" && s.SpeedLimit != "unknown"
}

func (s *SpeedLimitData) SpeedLimitInt() int {
	if !s.HasSpeedLimit() {
		return 0
	}
	limit, _ := strconv.Atoi(s.SpeedLimit)
	return limit
}

type TripData struct {
	StartTime    time.Time
	Distance     float64 // meters
	TotalSpeed   float64 // for averaging
	SpeedSamples int
}

func (t *TripData) Duration() time.Duration {
	if t.StartTime.IsZero() {
		return 0
	}
	return time.Since(t.StartTime)
}

func (t *TripData) AverageSpeed() float64 {
	if t.SpeedSamples == 0 {
		return 0
	}
	return t.TotalSpeed / float64(t.SpeedSamples)
}

func (t *TripData) DistanceKm() float64 {
	return t.Distance / 1000.0
}

type SettingsData struct {
	Theme              string // "dark", "light", "auto"
	Language           string // "en", "de"
	Mode               string // "speedometer", "navigation"
	ShowRawSpeed       bool
	ShowGps            string // "always", "active-or-error", "error", "never"
	ShowBluetooth      string
	ShowCloud          string
	ShowInternet       string
	ShowClock          string
	BlinkerStyle       string // "icon", "overlay"
	BatteryDisplayMode string // "percentage", "range"
	PowerDisplayMode   string // "kw", "amps"
	ValhallaURL        string

	DualBattery bool

	AlarmEnabled  bool
	AlarmHonk     bool
	AlarmDuration int // seconds

	AutoStandbySeconds int
	EnableHorn         string // "true", "false", "in-drive"
}

func DefaultSettings() *SettingsData {
	return &SettingsData{
		Theme:              "dark",
		Language:           "en",
		Mode:               "speedometer",
		ShowRawSpeed:       false,
		ShowGps:            "always",
		ShowBluetooth:      "always",
		ShowCloud:          "always",
		ShowInternet:       "always",
		ShowClock:          "always",
		DualBattery:        false,
		BlinkerStyle:       "icon",
		BatteryDisplayMode: "percentage",
		PowerDisplayMode:   "kw",
		AlarmEnabled:       false,
		AlarmHonk:          false,
		AlarmDuration:      10,
		EnableHorn:         "true",
	}
}

type OtaData struct {
	DbcStatus           string
	DbcUpdateVersion    string
	DbcDownloadProgress int
	DbcInstallProgress  int
	DbcError            string
	MdbStatus           string
	MdbUpdateVersion    string
	MdbDownloadProgress int
	MdbInstallProgress  int
	MdbError            string
}

func (o *OtaData) IsActive() bool {
	return (o.DbcStatus != "" && o.DbcStatus != "idle") ||
		(o.MdbStatus != "" && o.MdbStatus != "idle")
}
