package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/cmd/exporter/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/influxadapter"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
	"github.com/gin-gonic/gin"
)

var (
	repository  ports.Repository
	influxUri   string = os.Getenv("INFLUX_URI")
	influxToken string = os.Getenv("INFLUX_TOKEN")
	influxOrg   string = os.Getenv("INFLUX_ORG")
)

func init() {

	if len(influxUri) == 0 || len(influxToken) == 0 || len(influxOrg) == 0 {
		panic("Missing env variables!")
	}

	repository = influxadapter.NewInfluxDBRepository(influxUri, influxToken, influxOrg)
}

func main() {
	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/exporter/data", ExportData)
	r.Run("0.0.0.0:8000")
}

func ExportData(c *gin.Context) {

	start, stop, err := castQuery(c)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, domain.HttpError{
			Error: err.Error(),
		})
		return
	}

	data, err := repository.GetRecords(*start, *stop)

	if err != nil {
		log.Println("We fail miserable", err.Error())
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}

	c.Writer.Header().Set("Content-Disposition", "attachment; filename=sensor_reports.csv")
	c.Writer.Header().Set("Content-Type", "text/csv")
	writer := csv.NewWriter(c.Writer)
	header := []string{"Station", "Date", "IP", "AQI", "CO2", "Humidity", "Temperature", "TVOC"}
	if err := writer.Write(header); err != nil {
		c.String(http.StatusInternalServerError, "Error writing header")
		return
	}

	for station, reports := range data {
		for _, r := range reports {
			row := []string{
				station,
				r.Date.Format(time.RFC3339),
				r.StationIP,
				fmt.Sprintf("%.2f", r.AQI),
				fmt.Sprintf("%.2f", r.CO2),
				fmt.Sprintf("%.2f", r.Humidity),
				fmt.Sprintf("%.2f", r.Temperature),
				fmt.Sprintf("%.2f", r.Tvoc),
			}
			if err := writer.Write(row); err != nil {
				c.String(http.StatusInternalServerError, "Error writing row")
				return
			}
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		c.String(http.StatusInternalServerError, "Error flushing writer")
	}
}

func castQuery(c *gin.Context) (*time.Time, *time.Time, error) {
	var (
		startQuery = c.Query("from")
		stopQuery  = c.Query("until")
	)

	start, err := castStringToTime(startQuery)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("err start date is wrong"), err)
	}

	stop, err := castStringToTime(stopQuery)
	if err != nil {
		return nil, nil, errors.Join(fmt.Errorf("err stop date is wrong"), err)
	}

	return start, stop, nil

}

func castStringToTime(query string) (*time.Time, error) {

	querySplit := strings.Split(query, "-")
	// log.Println(querySplit)
	log.Println(query)

	if len(querySplit) != 3 {
		return nil, fmt.Errorf("err split should be YYYY-MM-DD, actal lenght %d", len(querySplit))
	}

	sYear, err := strconv.Atoi(querySplit[0])
	if err != nil {
		return nil, fmt.Errorf("err The year is on bad format on the date: \n %s", err.Error())

	}

	sMonth, err := strconv.Atoi(querySplit[1])
	if err != nil {
		return nil, fmt.Errorf("err the month is on bad format  on the date: \n %s", err.Error())

	}

	sDay, err := strconv.Atoi(querySplit[2])
	if err != nil {
		return nil, fmt.Errorf("err The day is on bad format on the date: \n %s", err.Error())

	}

	date := time.Date(
		sYear,
		time.Month(sMonth),
		sDay,
		0,
		0,
		0,
		0,
		time.UTC,
	)

	return &date, nil
}
