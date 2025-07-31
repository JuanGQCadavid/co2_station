package ports

import (
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
)

type Repository interface {
	GetRecords(start, stop time.Time) (map[string][]*domain.SensorReport, error)
}
