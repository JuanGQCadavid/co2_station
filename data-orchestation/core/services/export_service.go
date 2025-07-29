package services

import "github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"

type ExportService struct {
	repository ports.Repository
}

func NewExportService(repository ports.Repository) *ExportService {
	return &ExportService{
		repository: repository,
	}
}
