package domain

import "time"

type StationReport struct {
	IpAddress   string  `json:"ipAddress"`
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	Tvoc        float32 `json:"tvoc"`
	Co2         int     `json:"co2"`
	Aqi         int     `json:"aqi"`
}

type SensorReport struct {
	Date        time.Time // _time:2025-07-28 13:16:00 +0000 UTC
	StationIP   string    // ipAddress:192.168.0.62.
	AQI         float64   //_field:aqi.
	CO2         float64   // _field:co2
	Humidity    float64   //_field:humidity
	Temperature float64   //_field:temperature
	Tvoc        float64   //_field:tvoc
}
