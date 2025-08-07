package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/actionsdatabase"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/influxadapter"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/slackadapter"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/adapters/turtleboot"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/ports"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/services"
)

// The main idea is:
// // Follow the state machine.
// // Each step of the state machine code will be on the service
// // the main will only orchestate it and wait for it

// I should mock:
// // The sensors.
// // The turtle bot.

type LogLevel string

var (
	ERROR LogLevel = "ERROR"
	INFO  LogLevel = "INFO"
)

type State string

var (
	OnSensing         State = "OnSensing"
	OnMovingTo        State = "OnMovingTo"
	OnIntervention    State = "OnIntervention"
	OnBaseColdingDown State = "OnBaseColdingDown"
	OnError           State = "OnError"
	InitState         State = OnSensing
)

var (
	STATION_ID              = "base"
	WAITING_ON_INTERVENTION = 15 * time.Minute
	WAITING_ON_BASE         = 30 * time.Minute
)

type StateData struct {
	StationIP         string
	Error             error
	ErrorOnState      State
	AfterMovementGoTo State
}

var (
	service           *services.ControllerService
	slack             *slackadapter.SlackNotification
	actionsRepository *actionsdatabase.ActionsRepository
	theTurtle         *turtleboot.TurtleBoot
	repository        ports.Repository

	influxUri   string = os.Getenv("INFLUX_URI")
	influxToken string = os.Getenv("INFLUX_TOKEN")
	influxOrg   string = os.Getenv("INFLUX_ORG")

	slackTOken     string = os.Getenv("SLACK_TOKEN")
	slackChannelId string = os.Getenv("SLACK_CHANNEL_ID")

	turtleIPAddress string = os.Getenv("TURTLE_IP_ADDRESS")

	// host string, username string, password string, dbname string, port string
	actionsHost     string = os.Getenv("ACTIONS_HOST")
	actionsUsername string = os.Getenv("ACTIONS_USERNAME")
	actionsPassword string = os.Getenv("ACTIONS_PASSWORD")
	actionsDB       string = os.Getenv("ACTIONS_DB")
	actionsPort     string = os.Getenv("ACTIONS_PORT")

	// State Machine
	states map[State]func(*StateData)

	// on Transit varaibles
	maxAccomulativeErros = 5
	waitingTIme          = 5 * time.Second

	timeWindow = 24 * 31 * time.Hour // The last month
)

func init() {
	var (
		err error
	)

	if len(influxUri) == 0 || len(influxToken) == 0 || len(influxOrg) == 0 {
		panic("Missing Influx env variables!")
	}

	if len(slackTOken) == 0 || len(slackChannelId) == 0 {
		panic("Missing Slack env variables!")
	}

	if len(turtleIPAddress) == 0 {
		panic("Missing turtle ip address")
	}

	if len(actionsHost) == 0 || len(actionsUsername) == 0 || len(actionsPassword) == 0 || len(actionsDB) == 0 || len(actionsPort) == 0 {
		panic("Missing DB params")
	}

	slack = slackadapter.NewSlackNotification(slackTOken, slackChannelId)

	repository = influxadapter.NewInfluxDBRepository(influxUri, influxToken, influxOrg)
	actionsRepository = actionsdatabase.NewActionsDB(actionsHost, actionsUsername, actionsPassword, actionsDB, actionsPort)
	theTurtle, err = turtleboot.NewTurtleBoot(turtleIPAddress)

	if err != nil {
		log.Fatal("err while connecting to the turtle gRPC", err.Error())
	}

	service = services.NewControllerService(repository, theTurtle, actionsRepository)

	states = map[State]func(*StateData){
		OnSensing:         OnSensingFunc,
		OnMovingTo:        OnMovingToFunc,
		OnIntervention:    OnInterventionFunc,
		OnBaseColdingDown: OnBaseColdingDownFunc,
		OnError:           OnErrorFunc,
	}
}

func OnSensingFunc(_ *StateData) {
	log.Println(" ---------- OnSensingFunc ---------- ")

	theStation, err := service.FindAndSaveTheStation(timeWindow)

	if err == services.ErrSensorsNoSensing {
		log.Println("There is not data on the database, sleeping for 1 min", err.Error())
		time.Sleep(1 * time.Minute)
		states[InitState](nil)
		return
	}

	if err != nil {
		states[OnError](&StateData{
			Error:        err,
			ErrorOnState: OnSensing,
		})
		return
	}

	Notifiy(fmt.Sprintf("The slected was %s, with %v", theStation.StationIP, theStation.Indicator), INFO)

	states[OnMovingTo](&StateData{
		StationIP:         theStation.StationIP,
		AfterMovementGoTo: OnIntervention,
	})

}

func OnMovingToFunc(data *StateData) {
	log.Println(" ---------- OnMovingToFunc ---------- ")
	log.Printf("Data -> %+v\n", data)

	if err := service.InitMovement(data.StationIP); err != nil {
		log.Println("Err while initing the movement ", err.Error())

		states[OnError](&StateData{
			Error:        err,
			ErrorOnState: OnMovingTo,
		})
		return
	}
	if err := service.WaitUntilDoneOrError(data.StationIP, maxAccomulativeErros, waitingTIme); err != nil {
		log.Println("Err while waiting for the movement finish", err.Error())

		states[OnError](&StateData{
			Error:        err,
			ErrorOnState: OnMovingTo,
		})
		return
	}

	states[data.AfterMovementGoTo](&StateData{
		StationIP: data.StationIP,
	})

}

func OnInterventionFunc(data *StateData) {
	log.Println(" ---------- OnInterventionFunc ---------- ")
	log.Printf("Data -> %+v\n", data)

	ctx, cancel := context.WithTimeout(context.Background(), WAITING_ON_INTERVENTION)
	defer cancel()

	if err := service.Wait(ctx); err != nil {
		log.Println("Err while waiting in the st", err.Error())
		states[OnError](&StateData{
			Error:        err,
			ErrorOnState: OnIntervention,
		})
		return
	}

	states[OnMovingTo](&StateData{
		StationIP:         STATION_ID,
		AfterMovementGoTo: OnBaseColdingDown,
	})

}

func OnBaseColdingDownFunc(data *StateData) {
	log.Println(" ---------- OnBaseColdingDownFunc ---------- ")
	log.Printf("Data -> %+v\n", data)

	ctx, cancel := context.WithTimeout(context.Background(), WAITING_ON_BASE)
	defer cancel()

	if err := service.Wait(ctx); err != nil {
		log.Println("Err while waiting in the st", err.Error())
		states[OnError](&StateData{
			Error:        err,
			ErrorOnState: OnIntervention,
		})
		return
	}

	states[OnSensing](nil)
}

func OnErrorFunc(data *StateData) {
	log.Println(" ---------- OnErrorFunc ---------- ")
	log.Printf("Data -> %+v\n", data)

	Notifiy(
		fmt.Sprintf(
			"Turtlebot on error \n \tError:: %s \n \tState:%s \nMoving to base",
			data.Error.Error(),
			data.ErrorOnState,
		),
		ERROR,
	)

	// TODO - SHould go to the base or should Die and wait for someone starting this again?
	time.Sleep(12 * time.Hour) // Meanwhile just sleep

	states[OnMovingTo](&StateData{
		StationIP:         STATION_ID,
		AfterMovementGoTo: InitState,
	})

}

func Notifiy(msg string, level LogLevel) {
	log.Println(level, msg)
	slack.Send(msg, string(level))
}

func main() {
	Notifiy("Starting controller", INFO)
	states[InitState](nil)
}
