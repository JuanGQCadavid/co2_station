package utils

import (
	"math"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
)

func GenerateIndicator(reports []*domain.SensorReport) *domain.SensorReport {
	var (
		// Avegare
		co2Avg  float64
		vocsAvg float64
		tempAvg float64
		rhAvg   float64
		aqiAvg  float64

		// Station IP
		stationIP string

		// Reports length
		reportsLength = float64(len(reports))
	)

	// Getting Avg
	for _, rep := range reports {
		co2Avg += rep.CO2
		vocsAvg += rep.Tvoc
		tempAvg += rep.Temperature
		rhAvg += rep.Humidity
		aqiAvg += rep.AQI
		stationIP = rep.StationIP
	}

	co2Avg = co2Avg / reportsLength
	vocsAvg = vocsAvg / reportsLength
	tempAvg = tempAvg / reportsLength
	rhAvg = rhAvg / reportsLength
	aqiAvg = aqiAvg / reportsLength

	f := CalculateIndicator(co2Avg, vocsAvg, tempAvg, rhAvg, aqiAvg)

	return &domain.SensorReport{
		Indicator:   f,
		CO2:         co2Avg,
		StationIP:   stationIP,
		AQI:         aqiAvg,
		Humidity:    rhAvg,
		Temperature: tempAvg,
		Tvoc:        vocsAvg,
	}
}

func CalculateIndicator(co2Avg, vocsAvg, tempAvg, rhAvg, aqiAvg float64) float64 {
	var (
		w_1 = 0.30 // CO2
		w_2 = 0.25 // VOCs
		w_3 = 0.20 // Temperature
		w_4 = 0.15 // RH
		w_5 = 0.10 // AQI

		// Threashold
		// CO₂: TCO₂ = 1000 ppm (threshold for ventilation adequacy)
		t_1 float64 = 1000 // 1000 ppm

		// VOCs   (Volatile   Organic   Compounds):   TVOC   =   250   ppb   (threshold   forpollutant exposure)
		t_2 float64 = 250 //  250   pp

		//Temperature: Acceptable range = [18°C, 26°C] (thermal comfort boundaries)
		t_3 float64

		// Relative   Humidity   (RH):   Acceptable   range   =   [30%,   70%]
		t_4 float64

		// AQI   (Air   Quality   Index):   TAQI   =   3   (threshold   on   a   scale   from   1   to   5, representing moderate outdoor air quality impact)
		t_5 float64 = 3

		// Ranges
		t_3_range = [2]float64{
			18.0,
			26.0,
		}

		t_4_range = [2]float64{
			30.0,
			70.0,
		}
	)

	// Threasholds
	t_1 = math.Max(0, (co2Avg-t_1)/t_1)
	t_2 = math.Max(0, (vocsAvg-t_2)/t_2)
	t_3 = getScoreOnRange(tempAvg, t_3_range)
	t_4 = getScoreOnRange(rhAvg, t_4_range)
	t_5 = math.Max(0, (aqiAvg-t_5)/t_5)

	return w_1*t_1 + w_2*t_2 + w_3*t_3 + w_4*t_4 + w_5*t_5
}

func getScoreOnRange(val float64, ranges [2]float64) float64 {
	var (
		pivot float64 = 0
	)

	if val < ranges[0] {
		pivot = ranges[0]
	}

	if val > ranges[1] {
		pivot = ranges[1]
	}

	if pivot == 0 {
		return 0.0
	}

	return math.Max(0, (val-pivot)/pivot)
}
