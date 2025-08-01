package services

import (
	"context"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
)

type ControllerService struct {
	repository ports.Repository
}

func NewControllerService(repository ports.Repository) *ControllerService {
	return &ControllerService{
		repository: repository,
	}
}

type StationResult struct {
	StationIP string
	Indicator float64
}

// Sensing
func (svc *ControllerService) FindTheStation() (*StationResult, error) {
	// stations, err := svc.AnalyzeStationIndicator()
	return nil, nil
}

func (svc *ControllerService) AnalyzeStationIndicator() ([]*StationResult, error) {
	return nil, nil
}

// Moving to Intervation
func (svc *ControllerService) InitMovement(stationIP string) error {
	return nil
}

func (svc *ControllerService) WaitUntilDoneOrError(stationIP string) error {
	return nil
}

// Intervening
func (svc *ControllerService) Wait(ctx context.Context) error {
	return nil
}
