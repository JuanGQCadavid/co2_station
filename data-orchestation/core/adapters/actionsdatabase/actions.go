package actionsdatabase

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type ActionsRepository struct {
	db *gorm.DB
}

func NewActionsDB(host, username, password, dbname, port string) *ActionsRepository {

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", host, username, password, dbname, port)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	// Auto migrate the schema
	err = db.AutoMigrate(&ActionsDB{})
	if err != nil {
		log.Fatal("failed to migrate schema:", err)
	}

	return &ActionsRepository{
		db: db,
	}
}

func (repo *ActionsRepository) SaveAction(stationId string, airQualityScore float64) {
	result := repo.db.Create(&ActionsDB{
		StationID:       stationId,
		Datetime:        time.Now(),
		AirQualityScore: airQualityScore,
	})

	if result.Error != nil {
		log.Println("failed to insert user:", result.Error)
	}
}
