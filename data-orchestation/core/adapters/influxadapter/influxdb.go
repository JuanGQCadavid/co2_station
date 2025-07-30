package influxadapter

import (
	"context"
	"fmt"
	"log"
	"sort"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var (
	query = ` 
		from(bucket: "stations")
			|> range(start: %s, stop: %s)
			|> filter(fn: (r) => r["_measurement"] == "sensor")
			|> filter(fn: (r) => r["_field"] == "aqi" or r["_field"] == "co2" or r["_field"] == "humidity" or r["_field"] == "temperature" or r["_field"] == "tvoc")
			|> filter(fn: (r) => r["ipAddress"] == %q)
			|> filter(fn: (r) => r["topic"] == "report/drift")
			|> aggregateWindow(every: 1m, fn: mean, createEmpty: false)
			|> yield(name: "mean")
	`
)

type SensorReport struct {
	Date        time.Time // _time:2025-07-28 13:16:00 +0000 UTC
	StationIP   string    // ipAddress:192.168.0.62.
	AQI         float64   //_field:aqi.
	CO2         float64   // _field:co2
	Humidity    float64   //_field:humidity
	Temperature float64   //_field:temperature
	Tvoc        float64   //_field:tvoc
}

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

func (repo *InfluxDBRepository) GetRecords(start, stop time.Time, stationIP string) ([]*SensorReport, error) {
	queryAPI := repo.client.QueryAPI(repo.influxORG)
	thaQuery := fmt.Sprintf(query, start.Format(time.RFC3339), stop.Format(time.RFC3339), stationIP)

	result, err := queryAPI.Query(context.Background(), thaQuery)

	if err != nil {
		log.Printf("queryAPI query error: %s\n", err.Error())
		return nil, result.Err()
	}

	reports := make(map[time.Time]*SensorReport)

	// Casting
	for result.Next() {
		if reports[result.Record().Time()] == nil {
			reports[result.Record().Time()] = &SensorReport{
				Date:      result.Record().Time(),
				StationIP: result.Record().Values()["ipAddress"].(string),
			}
		}

		switch result.Record().Field() {
		case "aqi":
			reports[result.Record().Time()].AQI = result.Record().Value().(float64)
		case "co2":
			reports[result.Record().Time()].CO2 = result.Record().Value().(float64)
		case "humidity":
			reports[result.Record().Time()].Humidity = result.Record().Value().(float64)
		case "temperature":
			reports[result.Record().Time()].Temperature = result.Record().Value().(float64)
		case "tvoc":
			reports[result.Record().Time()].Tvoc = result.Record().Value().(float64)
		}
		fmt.Printf("Time: %s, Value: %v\n", result.Record().Time(), result.Record().Values())
	}

	if result.Err() != nil {
		log.Printf("Query error: %s\n", result.Err().Error())
		return nil, result.Err()
	}

	// Getting values

	records := make([]*SensorReport, 0, len(reports))
	for _, r := range reports {
		records = append(records, r)
	}

	// Sorting
	return repo.sortRecords(records), nil

}

func (repo *InfluxDBRepository) sortRecords(records []*SensorReport) []*SensorReport {

	sort.Slice(records, func(i, j int) bool {
		return records[i].Date.After(records[j].Date)
	})

	return records
}
