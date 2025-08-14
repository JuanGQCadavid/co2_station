package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/cmd/dashboard/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/actionsdatabase"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/influxadapter"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/services"
	"github.com/gin-gonic/gin"
)

var (
	//go:embed static/home.tmpl
	homeTmpl string
)

var (
	service           *services.ControllerService
	actionsRepository *actionsdatabase.ActionsRepository
	repository        ports.Repository

	influxUri   string = os.Getenv("INFLUX_URI")
	influxToken string = os.Getenv("INFLUX_TOKEN")
	influxOrg   string = os.Getenv("INFLUX_ORG")

	// host string, username string, password string, dbname string, port string
	actionsHost     string = os.Getenv("ACTIONS_HOST")
	actionsUsername string = os.Getenv("ACTIONS_USERNAME")
	actionsPassword string = os.Getenv("ACTIONS_PASSWORD")
	actionsDB       string = os.Getenv("ACTIONS_DB")
	actionsPort     string = os.Getenv("ACTIONS_PORT")

	// on Transit varaibles
	timeWindow = 5 * time.Minute //24 * 31 * time.Hour // The last month
)

func init() {
	var (
		err error
	)

	if len(influxUri) == 0 || len(influxToken) == 0 || len(influxOrg) == 0 {
		panic("Missing Influx env variables!")
	}

	if len(actionsHost) == 0 || len(actionsUsername) == 0 || len(actionsPassword) == 0 || len(actionsDB) == 0 || len(actionsPort) == 0 {
		panic("Missing DB params")
	}

	repository = influxadapter.NewInfluxDBRepository(influxUri, influxToken, influxOrg)
	actionsRepository = actionsdatabase.NewActionsDB(actionsHost, actionsUsername, actionsPassword, actionsDB, actionsPort)

	if err != nil {
		log.Fatal("err while connecting to the turtle gRPC", err.Error())
	}

	service = services.NewControllerService(repository, nil, actionsRepository)
}

func main() {

	r := gin.Default()
	r.LoadHTMLGlob("static/*")

	r.GET("/", HomePage)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/intervene", Intervene)

	r.Run("0.0.0.0:80")
}

func Intervene(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		log.Println(err.Error())
		return
	}

	var (
		station    = c.Request.Form.Get("Station")
		score, err = strconv.ParseFloat(c.Request.Form.Get("Score"), 64)
	)

	if len(station) == 0 {
		log.Println("Missing values: station: ", station, " score: ", score)
		c.AbortWithStatusJSON(http.StatusBadRequest, domain.HttpError{
			Error: "Missing values station or score",
		})
		return
	}

	if err != nil {
		log.Println("Could not cast score: ", err.Error())
		log.Println(c.Request.Form.Get("Score"))
		c.AbortWithStatusJSON(http.StatusBadRequest, domain.HttpError{
			Error: fmt.Sprintf("Could not cast score: %s", err.Error()),
		})
		return
	}

	if err := service.SaveStation(station, score); err != nil {
		log.Println("Err while saving the station", err.Error())
		c.AbortWithStatusJSON(http.StatusInternalServerError, domain.HttpError{
			Error: fmt.Sprintf("Could save the data, err %s", err.Error()),
		})
		return
	}

	c.HTML(http.StatusOK, "intervene.tmpl", gin.H{
		"Station": station,
		"Score":   fmt.Sprintf("%.2f", score),
	})
}

func HomePage(c *gin.Context) {
	results, err := service.AnalyzeStationIndicatorV2(timeWindow)

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, domain.HttpError{
			Error: err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "home.tmpl", gin.H{
		"Sensors": results,
	})
}
