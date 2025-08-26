package services

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/actionsdatabase"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/turtleboot"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/utils"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/pb"
)

var (
	ErrTurtleDie        = errors.New("err turtle needs a human")
	ErrSensorsNoSensing = errors.New("err Sensors are not reporting!! check sensors")
)

type ControllerService struct {
	repository   ports.Repository
	turtleBoot   *turtleboot.TurtleBoot
	dbRepository *actionsdatabase.ActionsRepository
}

func NewControllerService(repository ports.Repository, turtleBoot *turtleboot.TurtleBoot, dbRepository *actionsdatabase.ActionsRepository) *ControllerService {
	return &ControllerService{
		repository:   repository,
		turtleBoot:   turtleBoot,
		dbRepository: dbRepository,
	}
}

// Sensing
func (svc *ControllerService) FindTheStation(timeWindow time.Duration) (*domain.StationResult, error) {
	stations, err := svc.AnalyzeStationIndicator(timeWindow)

	if err != nil {
		log.Println("error analyzing the stations ", err.Error())
		return nil, err
	}

	if len(stations) == 0 {
		return nil, ErrSensorsNoSensing
	}

	sort.Slice(stations, func(i, j int) bool {
		return stations[i].Indicator >= stations[j].Indicator
	})

	return stations[0], nil
}

func (svc *ControllerService) FindAndSaveTheStation(timeWindow time.Duration) (*domain.StationResult, error) {
	theSation, err := svc.FindTheStation(timeWindow)

	if err != nil {
		return nil, err
	}

	id, err := svc.dbRepository.SaveAction(theSation.StationIP, theSation.Indicator)

	if err != nil {
		return theSation, err
	}

	theSation.Id = id
	return theSation, nil
}

func (svc *ControllerService) SaveStation(stationIP string, score float64) (*domain.StationResult, error) {

	var result = &domain.StationResult{
		StationIP: stationIP,
		Indicator: score,
	}
	id, err := svc.dbRepository.SaveAction(stationIP, score)

	if err != nil {
		return nil, err
	}
	result.Id = id
	return result, nil
}

func (svc *ControllerService) StopIntervention(interventionId string) error {
	interId, err := strconv.ParseUint(interventionId, 10, 64)

	if err != nil {
		fmt.Println("Error parsing:", err)
		return fmt.Errorf("Error parsing the id %s", err.Error())
	}

	svc.dbRepository.StopIntervention(interId, time.Now())
	return nil
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
			Indicator: utils.GenerateIndicator(reportsOfStations).Indicator,
		})
	}
	return stationResults, nil
}

func (svc *ControllerService) AnalyzeStationIndicatorV2(timeWindow time.Duration) ([]*domain.SensorReport, error) {
	reports, err := svc.repository.GetRecords(time.Now().Add(-timeWindow), time.Now())

	if err != nil {
		log.Println("Err from repository ", err.Error())
		return nil, err
	}

	var (
		stationResults = make([]*domain.SensorReport, 0, len(reports))
	)

	for _, reportsOfStations := range reports {
		stationResults = append(stationResults, utils.GenerateIndicator(reportsOfStations))
	}

	sort.Slice(stationResults, func(i, j int) bool {
		return stationResults[i].Indicator >= stationResults[j].Indicator
	})

	return stationResults, nil
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
