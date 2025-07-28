package main

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
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
		log.Println("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		log.Println("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}
}

func OnClientError(err error) {
	log.Println("client error: %s\n", err)
}

func main2() {
	// App will run until cancelled by user (e.g. ctrl-c)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{u},
		KeepAlive:  20, // Keepalive message should be sent every 20 seconds

		// CleanStartOnInitialConnection defaults to false. Setting this to true will clear the session on the first connection.
		CleanStartOnInitialConnection: false,

		// SessionExpiryInterval - Seconds that a session will survive after disconnection.
		// It is important to set this because otherwise, any queued messages will be lost if the connection drops and
		// the server will not queue messages while it is down. The specific setting will depend upon your needs
		// (60 = 1 minute, 3600 = 1 hour, 86400 = one day, 0xFFFFFFFE = 136 years, 0xFFFFFFFF = don't expire)
		SessionExpiryInterval: 60,

		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			log.Println("mqtt connection up")
			// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
			// the connection drops)
			if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
				Subscriptions: []paho.SubscribeOptions{
					{Topic: topic, QoS: 1},
				},
			}); err != nil {
				fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
			}

			fmt.Println("mqtt subscription made")
		},

		OnConnectError: func(err error) {
			log.Printf("error whilst attempting connection: %s\n", err)
			log.Panic(err.Error())
		},

		// eclipse/paho.golang/paho provides base mqtt functionality, the below config will be passed in for each connection
		ClientConfig: paho.ClientConfig{
			// If you are using QOS 1/2, then it's important to specify a client id (which must be unique)
			ClientID: clientID,

			// OnPublishReceived is a slice of functions that will be called when a message is received.
			// You can write the function(s) yourself or use the supplied Router
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					fmt.Printf("received message on topic %s; body: %s (retain: %t)\n", pr.Packet.Topic, pr.Packet.Payload, pr.Packet.Retain)
					return true, nil
				}},
			OnClientError: func(err error) { fmt.Printf("client error: %s\n", err) },
			OnServerDisconnect: func(d *paho.Disconnect) {
				if d.Properties != nil {
					fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
				} else {
					fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
				}
			},
		},
	}

	c, err := autopaho.NewConnection(ctx, cliCfg) // starts process; will reconnect until context cancelled
	if err != nil {
		panic(err)
	}
	// Wait for the connection to come up
	if err = c.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	ticker := time.NewTicker(time.Second)
	msgCount := 0
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			msgCount++
			// Publish a test message (use PublishViaQueue if you don't want to wait for a response)
			if _, err = c.Publish(ctx, &paho.Publish{
				QoS:     1,
				Topic:   topic,
				Payload: []byte("TestMessage: " + strconv.Itoa(msgCount)),
			}); err != nil {
				if ctx.Err() == nil {
					panic(err) // Publish will exit when context cancelled or if something went wrong
				}
			}
			continue
		case <-ctx.Done():
		}
		break
	}

	fmt.Println("signal caught - exiting")
	<-c.Done() // Wait for clean shutdown (cancelling the context triggered the shutdown)
}
