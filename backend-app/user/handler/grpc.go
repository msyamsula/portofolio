package handler

import (
	"context"
	"log"
	"net"

	pb "github.com/msyamsula/portofolio/backend-app/user/proto"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"

	"google.golang.org/grpc"
)

// implement the server
type grpcHandler struct {
	pb.UnimplementedExampleServiceServer
	internalToken internaltoken.InternalToken
}

// handler for SayHello
func (s *grpcHandler) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	name := req.GetName()
	s.internalToken.ValidateToken(ctx, "")
	return &pb.HelloResponse{
		Message: "Hello, " + name + "!",
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterExampleServiceServer(s, &grpcHandler{})

	log.Println("gRPC server running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
