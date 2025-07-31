package influxadapter

import (
	"fmt"
	"log"
	"testing"
	"time"
)

func TestFetch(t *testing.T) {

	var (
		influxURI   = "http://localhost:8086"
		influxToken = "my-super-secret-auth-token"
		influxORG   = "ut"
	)

	start := time.Date(2025, 7, 27, 0, 0, 0, 0, time.UTC)
	stop := time.Date(2025, 7, 29, 0, 0, 0, 0, time.UTC)

	influxRepo := NewInfluxDBRepository(
		influxURI, influxToken, influxORG,
	)

	stations, err := influxRepo.GetRecords(start, stop)

	if err != nil {
		log.Fatalln(err.Error())
	}

	for ip, stationReports := range stations {
		fmt.Println(ip)
		for _, report := range stationReports {
			fmt.Printf(" \t %+v \n", report)
		}

	}
}
