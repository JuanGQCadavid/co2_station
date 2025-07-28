package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	"github.com/eclipse/paho.golang/paho"
)

type ExtractService struct {
	topic string
}

func NewExtractService(topic string) *ExtractService {
	return &ExtractService{
		topic: topic,
	}
}

func (srv *ExtractService) Handle(msg paho.PublishReceived) (bool, error) {
	var (
		payload   *domain.StationReport = &domain.StationReport{}
		toPublish []byte
	)

	if err := json.Unmarshal(msg.Packet.Payload, payload); err != nil {
		log.Println("Error while unmarhsaling", err.Error())
		log.Println("Payload: ", string(msg.Packet.Payload))
		return false, err
	}

	toPublish, err := json.Marshal(srv.applyCorrections(payload))

	if err != nil {
		log.Println("Error while marhsaling", err.Error())
		return false, err
	}

	if _, err = msg.Client.Publish(context.Background(), &paho.Publish{
		QoS:     msg.Packet.QoS,
		Topic:   srv.topic,
		Payload: toPublish,
	}); err != nil {
		log.Println("Error while publishing", err.Error())
		return false, err
	}

	return true, nil
}

func (srv *ExtractService) applyCorrections(payload *domain.StationReport) *domain.StationReport {
	payload.Temperature = payload.Temperature + 4.43
	payload.Humidity = payload.Humidity - 12.27
	return payload
}
