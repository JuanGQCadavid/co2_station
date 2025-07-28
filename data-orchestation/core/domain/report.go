package domain

type StationReport struct {
	IpAddress   string  `json:"ipAddress"`
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	Tvoc        float32 `json:"tvoc"`
	Co2         int     `json:"co2"`
	Aqi         int     `json:"aqi"`
}
