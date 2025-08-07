package services

import (
	"context"
	"errors"
	"log"
	"math"
	"sort"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/turtleboot"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/pb"
)

var (
	timeWindow   = 24 * 31 * time.Hour // The last month
	ErrTurtleDie = errors.New("err turtle needs a human")
)

type ControllerService struct {
	repository ports.Repository
	turtleBoot *turtleboot.TurtleBoot
}

func NewControllerService(repository ports.Repository, turtleBoot *turtleboot.TurtleBoot) *ControllerService {
	return &ControllerService{
		repository: repository,
		turtleBoot: turtleBoot,
	}
}

// Sensing
func (svc *ControllerService) FindTheStation() (*domain.StationResult, error) {
	stations, err := svc.AnalyzeStationIndicator(timeWindow)

	if err != nil {
		log.Println("error analyzing the stations ", err.Error())
		return nil, err
	}

	if len(stations) == 0 {
		return nil, errors.New("err Sensors are not reporting!! check sensors")
	}

	sort.Slice(stations, func(i, j int) bool {
		return stations[i].Indicator >= stations[j].Indicator
	})

	return stations[0], nil
}

func (svc *ControllerService) AnalyzeStationIndicator(timeWindow time.Duration) ([]*domain.StationResult, error) {
	reports, err := svc.repository.GetRecords(time.Now().Add(-timeWindow), time.Now())

	if err != nil {
		log.Println("Err from repository ", err.Error())
		return nil, err
	}

	var (
		stationResults = make([]*domain.StationResult, 0, len(reports))
	)

	for station, reportsOfStations := range reports {
		stationResults = append(stationResults, &domain.StationResult{
			StationIP: station,
			Indicator: svc.GenerateIndicator(reportsOfStations),
		})
	}
	return stationResults, nil
}

func (svc *ControllerService) GenerateIndicator(reports []*domain.SensorReport) float64 {
	var (
		w_1 = 0.30 // CO2
		w_2 = 0.25 // VOCs
		w_3 = 0.20 // Temperature
		w_4 = 0.15 // RH
		w_5 = 0.10 // AQI

		// Avegare
		co2Avg  float64
		vocsAvg float64
		tempAvg float64
		rhAvg   float64
		aqiAvg  float64

		// Reports length
		reportsLength = float64(len(reports))

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

	// Getting Avg
	for _, rep := range reports {
		co2Avg += rep.CO2
		vocsAvg += rep.Tvoc
		tempAvg += rep.Temperature
		rhAvg += rep.Humidity
		aqiAvg += rep.AQI
	}

	co2Avg = co2Avg / reportsLength
	vocsAvg = vocsAvg / reportsLength
	tempAvg = tempAvg / reportsLength
	rhAvg = rhAvg / reportsLength
	aqiAvg = aqiAvg / reportsLength

	// Threasholds

	t_1 = math.Max(0, (co2Avg-t_1)/t_1)
	t_2 = math.Max(0, (vocsAvg-t_2)/t_2)
	t_3 = svc.getScoreOnRange(tempAvg, t_3_range)
	t_4 = svc.getScoreOnRange(rhAvg, t_4_range)
	t_5 = math.Max(0, (aqiAvg-t_5)/t_5)

	return w_1*t_1*co2Avg + w_2*t_2*vocsAvg + w_3*t_3*tempAvg + w_4*t_4*rhAvg + w_5*t_5*aqiAvg
}

func (svc *ControllerService) getScoreOnRange(val float64, ranges [2]float64) float64 {
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

// Moving to Intervation
func (svc *ControllerService) InitMovement(stationIP string) error {
	return svc.turtleBoot.MoveToStation(stationIP)
}

func (svc *ControllerService) WaitUntilDoneOrError(stationIP string, maxAccomualtiveErros int, waitTime time.Duration) error {
	if err := svc.InitMovement(stationIP); err != nil {
		return err
	}

	var (
		accomulativeErros = 0
		accomulativeError = errors.New("err while wating for report")
	)

	for accomulativeErros < maxAccomualtiveErros {
		state, err := svc.turtleBoot.ReportStatus()

		if err != nil {
			log.Println("Err while geting status from bot ", err.Error())
			accomulativeErros += 1
			accomulativeError = errors.Join(accomulativeError, err)
		}

		switch state.State {
		case pb.AgentState_IN_TRANSIT:
			log.Println("Turtle in transit, sleeping")
			time.Sleep(waitTime)
		case pb.AgentState_IN_STATION, pb.AgentState_IN_BASE:
			log.Println("Turtle reached station")
			return nil
		case pb.AgentState_ON_ERROR:
			log.Println("Ups turtle needs a human!")
			return ErrTurtleDie
		}
	}

	return accomulativeError
}

// Intervening
func (svc *ControllerService) Wait(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
