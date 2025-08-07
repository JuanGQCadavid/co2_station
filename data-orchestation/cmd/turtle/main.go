package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/JuanGQCadavid/co2_station/data-orchestation/pb"
	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedAgentControlServer
	state   pb.AgentState
	station string
}

func (srv *server) MoveToStation(ctx context.Context, req *pb.MoveCommand) (*pb.Response, error) {
	srv.state = pb.AgentState_IN_TRANSIT
	srv.station = ""
	go func(req *pb.MoveCommand) {
		log.Println("Command recived, now I will walk sleeping")
		time.Sleep(1 * time.Minute)

		log.Println("Here am I!")
		if req.StationId == "base" {
			srv.state = pb.AgentState_IN_BASE
			srv.station = "base"
			return
		}

		srv.state = pb.AgentState_IN_STATION
		srv.station = req.StationId
	}(req)
	return &pb.Response{Ok: true}, nil
}
func (srv *server) ReportStatus(context.Context, *pb.Empty) (*pb.Status, error) {
	return &pb.Status{
		State:        srv.state,
		StationId:    srv.station,
		BatteryState: 0.5,
	}, nil
}
func main() {
	lis, err := net.Listen("tcp", ":50051")

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAgentControlServer(s, &server{
		state: pb.AgentState_IN_BASE,
	})
	log.Println("gRPC server is running on port 50051")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
