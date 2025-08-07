package services

import (
	"log"
	"testing"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/influxadapter"
	"github.com/stretchr/testify/assert"
)

func TestScores(t *testing.T) {
	var (
		influxURI   = "http://localhost:8086"
		influxToken = "my-super-secret-auth-token"
		influxORG   = "ut"
	)

	influxRepo := influxadapter.NewInfluxDBRepository(
		influxURI, influxToken, influxORG,
	)

	theController := NewControllerService(influxRepo, nil)

	results, err := theController.FindTheStation()

	if err != nil {
		log.Println("Fuck we fail", err.Error())
	}
	assert.Nil(t, err, "Err is not nill! ")

	log.Printf("%+v\n", results)

}
