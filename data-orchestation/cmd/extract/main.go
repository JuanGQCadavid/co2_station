package main

import (
	"context"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/services"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

const (
	clientID      string = "GoExtract"
	topic         string = "station/report"
	mqtt_env_name string = "mqtt_uri"
	respond_topic string = "report/drift"
)

var (
	mqtt_uri        string = "mqtt://mosquitto:1883"
	u               *url.URL
	exctractServive *services.ExtractService
)

func init() {
	var (
		err error
	)
	mqtt, ok := os.LookupEnv(mqtt_env_name)

	if ok {
		mqtt_uri = mqtt
	}

	u, err = url.Parse(mqtt_uri)
	if err != nil {
		panic(err)
	}

	exctractServive = services.NewExtractService(
		respond_topic,
	)

}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     20,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60, // Seconds that a session will survive after disconnection.
		OnConnectionUp:                OnConnectionUp,
		OnConnectError:                OnConnectError,
		ClientConfig: paho.ClientConfig{
			ClientID: clientID,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				exctractServive.Handle,
			},
			OnClientError:      OnClientError,
			OnServerDisconnect: OnServerDisconnect,
		},
	}

	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		panic(err)
	}

	// Waiting for an interruption to stop the process
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	signal.Notify(sig, syscall.SIGTERM)
	<-sig

	log.Println("signal caught - exiting")

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	_ = cm.Disconnect(ctx)
	log.Println("shutdown complete")

}

func OnConnectError(err error) {
	log.Printf("error whilst attempting connection: %s\n", err)
	log.Panic(err.Error())
}

func OnConnectionUp(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	log.Println("mqtt connection up")

	if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{
			{
				Topic: topic,
				QoS:   2,
			},
		},
	}); err != nil {
		log.Panic("failed to subscribe!. ", err.Error())
	}

	log.Println("mqtt subscription made")
}

func OnServerDisconnect(d *paho.Disconnect) {
	if d.Properties != nil {
		log.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		log.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}
}

func OnClientError(err error) {
	log.Printf("client error: %s\n", err)
}
