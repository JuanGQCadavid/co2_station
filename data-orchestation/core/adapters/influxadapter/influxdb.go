package influxadapter

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	query = ` 
		from(bucket: "stations")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "sensor")
			|> filter(fn: (r) => r["_field"] == "aqi" or r["_field"] == "co2" or r["_field"] == "humidity" or r["_field"] == "temperature" or r["_field"] == "tvoc")
			|> filter(fn: (r) => r["topic"] == "report/drift")
			|> aggregateWindow(every: 1m, fn: mean, createEmpty: false)
			|> yield(name: "mean")
	`
)

type InfluxDBRepository struct {
	client      influxdb2.Client
	influxURI   string
	influxToken string
	influxORG   string
}

func NewInfluxDBRepository(influxURI, influxToken, influxORG string) *InfluxDBRepository {
	return &InfluxDBRepository{
		client:      influxdb2.NewClient(influxURI, influxToken),
		influxURI:   influxURI,
		influxToken: influxToken,
		influxORG:   influxORG,
	}
}

func (repo *InfluxDBRepository) GetRecords(start, stop time.Time) (map[string][]*domain.SensorReport, error) {
	queryAPI := repo.client.QueryAPI(repo.influxORG)
	thaQuery := fmt.Sprintf(query, start.Format(time.RFC3339), stop.Format(time.RFC3339))

	result, err := queryAPI.Query(context.Background(), thaQuery)

	if err != nil {
		log.Printf("queryAPI query error: %s\n", err.Error())
		return nil, result.Err()
	}

	reports := make(map[string]map[time.Time]*domain.SensorReport)

	// Casting
	for result.Next() {
		var (
			ip = result.Record().Values()["ipAddress"].(string)
		)

		if reports[ip] == nil {
			reports[ip] = make(map[time.Time]*domain.SensorReport)
		}

		if reports[ip][result.Record().Time()] == nil {
			reports[ip][result.Record().Time()] = &domain.SensorReport{
				Date:      result.Record().Time(),
				StationIP: result.Record().Values()["ipAddress"].(string),
			}
		}

		switch result.Record().Field() {
		case "aqi":
			reports[ip][result.Record().Time()].AQI = result.Record().Value().(float64)
		case "co2":
			reports[ip][result.Record().Time()].CO2 = result.Record().Value().(float64)
		case "humidity":
			reports[ip][result.Record().Time()].Humidity = result.Record().Value().(float64)
		case "temperature":
			reports[ip][result.Record().Time()].Temperature = result.Record().Value().(float64)
		case "tvoc":
			reports[ip][result.Record().Time()].Tvoc = result.Record().Value().(float64)
		}
		// fmt.Printf("Time: %s, Value: %v\n", result.Record().Time(), result.Record().Values())
	}

	if result.Err() != nil {
		log.Printf("Query error: %s\n", result.Err().Error())
		return nil, result.Err()
	}

	// Getting values

	var (
		sensorsData = make(map[string][]*domain.SensorReport)
	)

	for ip, stationReport := range reports {

		values := make([]*domain.SensorReport, 0, len(stationReport))
		for _, k := range stationReport {
			values = append(values, k)
		}

		sensorsData[ip] = repo.sortRecords(values)
	}

	// Sorting
	return sensorsData, nil

}

func (repo *InfluxDBRepository) sortRecords(records []*domain.SensorReport) []*domain.SensorReport {

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.After(records[j].Date)
	})

	return records
}
