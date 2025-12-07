package handler

import (
	"context"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/msyamsula/portofolio/backend-app/pkg/logger"
	pb "github.com/msyamsula/portofolio/backend-app/user/proto"
	internaltoken "github.com/msyamsula/portofolio/backend-app/user/service/internal-token"
	"go.opentelemetry.io/otel"
)

// implement the server
type grpcHandler struct {
	pb.UnimplementedExampleServiceServer
	internalToken internaltoken.InternalToken
}

// handler for SayHello
func (s *grpcHandler) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	var span trace.Span
	_, span = otel.Tracer("").Start(ctx, "grpcHandler.SayHello")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	name := req.GetName()
	var userData internaltoken.UserData
	userData, err = s.internalToken.ValidateToken(ctx, "")
	if err != nil {
		logger.Logger.Error(err.Error())
	}
	logger.Logger.Infof("validated user data: %+v", userData)
	return &pb.HelloResponse{
		Message: "Hello, " + name + "!",
	}, nil
}
