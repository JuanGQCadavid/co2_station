package main

import (
	"context"
	"encoding/json"
	"log"
	"math/rand"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

// I should simulate:
// 1. Sending data to queue - 3 stations
// 2. Turte boot, recives, acts and works pretty good
// 3. Turtle boot, recives and crash after 1 mins

func main() {
	var (
		ctx, stop = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		sensors   = []string{
			"192.168.0.145",
			"192.168.0.146",
			"192.168.0.147",
		}
		mqttDNS = "mqtt://localhost:1883"
	)
	defer stop()

	u, err := url.Parse(mqttDNS)
	if err != nil {
		panic(err)
	}

	for _, sensor := range sensors {
		go SensorSimulation(ctx, sensor, u)
	}

	// Waiting for an interruption to stop the process
	<-ctx.Done()
}

func SensorSimulation(ctx context.Context, ipAddress string, u *url.URL) {
	var (
		randTimeToSleep = 3
		cliCfg          = autopaho.ClientConfig{
			ServerUrls:                    []*url.URL{u},
			KeepAlive:                     20,
			ConnectRetryDelay:             5 * time.Second,
			CleanStartOnInitialConnection: false,
			SessionExpiryInterval:         60, // Seconds that a session will survive after disconnection.
			OnConnectError: func(err error) {
				log.Fatal("Connection err:", err)
			},

			ClientConfig: paho.ClientConfig{
				ClientID: ipAddress,
				OnClientError: func(err error) {
					log.Println("Client error:", err)
				},
				OnServerDisconnect: func(d *paho.Disconnect) {
					log.Println("Server disconnected:", d.Properties.ReasonString)
				},
			},
		}
	)

	mqttClient, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		log.Fatalf("Failed to create autopaho connection: %v", err)
	}

	if err := mqttClient.AwaitConnection(ctx); err != nil {
		log.Fatalf("Failed to connect to broker: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			// Context cancelled or expired
			log.Println("Context done:", ctx.Err())
			return
		default:
			sendSensorSignal(ctx, ipAddress, mqttClient)
			time.Sleep(time.Second * time.Duration(randTimeToSleep))
		}
	}
}

type StationReport struct {
	IpAddress   string  `json:"ipAddress"`
	Humidity    float32 `json:"humidity"`
	Temperature float32 `json:"temperature"`
	Tvoc        float32 `json:"tvoc"`
	Co2         int     `json:"co2"`
	Aqi         int     `json:"aqi"`
}

func sendSensorSignal(ctx context.Context, ipAddress string, mqttClient *autopaho.ConnectionManager) {
	log.Println("Station: ", ipAddress, " Sending data.")

	data := StationReport{
		IpAddress:   ipAddress,
		Co2:         rand.Intn(10000),
		Humidity:    rand.Float32() * 100,
		Temperature: rand.Float32() * 100,
		Tvoc:        rand.Float32() * 1000,
		Aqi:         rand.Intn(6),
	}

	payload, err := json.Marshal(&data)

	if err != nil {
		log.Println("ERRORR!! Station ", ipAddress, " on marshal err ", err.Error())
	}

	if _, err := mqttClient.Publish(ctx, &paho.Publish{
		Topic:   "station/report",
		QoS:     1,
		Payload: payload,
	}); err != nil {
		log.Println("ERRORR!! Station ", ipAddress, " on publish err: ", err.Error())
	}

}
