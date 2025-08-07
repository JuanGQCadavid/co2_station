package actionsdatabase

import (
	"time"

	"gorm.io/gorm"
)

type ActionsDB struct {
	ID              uint   `gorm:"primaryKey"`
	StationID       string `gorm:"index"`
	Datetime        time.Time
	AirQualityScore float64

	// GORM Variables
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
