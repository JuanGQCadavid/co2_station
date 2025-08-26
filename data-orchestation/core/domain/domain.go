package domain

import (
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/pb"
)

type StationReport struct {
	IpAddress   string  `json:"ipAddress"`
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	Tvoc        float32 `json:"tvoc"`
	Co2         int     `json:"co2"`
	Aqi         int     `json:"aqi"`
}

type SensorReport struct {
	Date        time.Time `json:"-"`
	StationIP   string    `json:"ipAddress"`
	AQI         float64   `json:"aqi"`
	CO2         float64   `json:"co2"`
	Humidity    float64   `json:"humidity"`
	Temperature float64   `json:"temperature"`
	Tvoc        float64   `json:"tvoc"`
	Indicator   float64   `json:"qualityIndicator"`
}

type StationResult struct {
	StationIP string
	Indicator float64
}

type TurtleState struct {
	Battery    float32
	State      pb.AgentState
	StattionId string
}
