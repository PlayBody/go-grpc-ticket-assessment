package main

import (
	"flag"
	"fmt"
	"github.com/playbody/train-ticket-service/proto"
	server "github.com/playbody/train-ticket-service/server"
	"google.golang.org/grpc"
	"log"
	"net"
)

var (
	port       = flag.Int("port", 50051, "The server port")
	configPath = flag.String("config", "config.yaml", "The config file to import init values")
)

func main() {
	err := server.SConfig.InitConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to init config: %v", err)
	}
	flag.Parse()
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer(grpc.UnaryInterceptor(server.ParseJWTMiddleware))
	var trainServer = &server.TrainServer{
		Conf: &server.SConfig.Train,
	}
	trainServer.InitServer()
	proto.RegisterTrainServiceServer(s, trainServer)
	log.Printf("server listening at %v", listener.Addr())
	if err := s.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
