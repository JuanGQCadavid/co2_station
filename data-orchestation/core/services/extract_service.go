package services

import (
	"context"
	"encoding/json"
	"log"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/utils"
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
		payload   *domain.SensorReport = &domain.SensorReport{}
		toPublish []byte
	)

	if err := json.Unmarshal(msg.Packet.Payload, payload); err != nil {
		log.Println("Error while unmarhsaling", err.Error())
		log.Println("Payload: ", string(msg.Packet.Payload))
		return false, err
	}

	toPublish, err := json.Marshal(
		srv.calcualteQualityIndicator(
			srv.applyCorrections(payload),
		),
	)

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

func (srv *ExtractService) applyCorrections(payload *domain.SensorReport) *domain.SensorReport {
	payload.Temperature = payload.Temperature - 4.43
	payload.Humidity = payload.Humidity + 12.27
	return payload
}

func (srv *ExtractService) calcualteQualityIndicator(payload *domain.SensorReport) *domain.SensorReport {
	payload.Indicator = utils.CalculateIndicator(payload.CO2, payload.Tvoc, payload.Temperature, payload.Humidity, payload.AQI)
	return payload
}
