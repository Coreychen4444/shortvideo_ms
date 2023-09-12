package handler

import (
	"flag"
	"log"

	pb "github.com/Coreychen4444/shortvideo"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	user_addr  = flag.String("addr", "user_service:50051", "the address to connect to")
	video_addr = flag.String("addr", "video_service:50052", "the address to connect to")
)

func UserClient() *pb.UserServiceClient {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*user_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewUserServiceClient(conn)
	return &c
}

func VideoClient() *pb.VideoServiceClient {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*video_addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewVideoServiceClient(conn)
	return &c
}

type ServiceHandler struct {
	uc pb.UserServiceClient
	vc pb.VideoServiceClient
}

func NewServiceHandler(uc *pb.UserServiceClient, vc *pb.VideoServiceClient) *ServiceHandler {
	return &ServiceHandler{uc: *uc, vc: *vc}
}
