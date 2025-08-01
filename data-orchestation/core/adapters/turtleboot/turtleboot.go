package turtleboot

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/core/domain"
	"github.com/JuanGQCadavid/co2_station/data-orchestation/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type TurtleBoot struct {
	address string
	client  pb.AgentControlClient
}

func NewTurtleBoot(address string) (*TurtleBoot, error) {

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))

	client := pb.NewAgentControlClient(conn)

	if err != nil {
		log.Printf("did not connect: %v \n", err)
		return nil, err
	}

	return &TurtleBoot{
		address: address,
		client:  client,
	}, nil
}

func (t *TurtleBoot) MoveToStation(stationID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := t.client.MoveToStation(ctx, &pb.MoveCommand{
		StationId: stationID,
	})

	if err != nil {
		log.Println("Error by calling the turtle boot on MoveToStation ", err.Error())
		return err
	}

	if !resp.Ok {
		return fmt.Errorf("err MoveToStation - client respond: %s", resp.OnError)
	}

	return nil
}

func (t *TurtleBoot) ReportStatus() (*domain.TurtleState, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	resp, err := t.client.ReportStatus(ctx, &pb.Empty{})

	if err != nil {
		log.Println("Error by calling the turtle boot on ReportStatus ", err.Error())
		return nil, err
	}

	return &domain.TurtleState{
		Battery:    resp.BatteryState,
		State:      resp.State,
		StattionId: resp.StationId,
	}, nil
}
