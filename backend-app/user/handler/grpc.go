package handler

import (
	"context"

	pb "github.com/msyamsula/portofolio/backend-app/user/proto"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
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
