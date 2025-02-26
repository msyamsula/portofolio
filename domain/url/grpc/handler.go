package grpc

import (
	"context"
	"log"

	"github.com/msyamsula/portofolio/domain/url/grpc/pb"
	urlsvc "github.com/msyamsula/portofolio/domain/url/service"
)

type UrlHandler struct {
	pb.UnimplementedUrlShortenerServer

	urlService *urlsvc.Service
}

type Dependencies struct {
	UrlService *urlsvc.Service
}

func New(dep Dependencies) *UrlHandler {
	if dep.UrlService == nil {
		log.Fatal("empty url service")
		return nil
	}
	return &UrlHandler{
		urlService: dep.UrlService,
	}
}

func (h *UrlHandler) GetLongUrl(ctx context.Context, data *pb.UrlRequest) (*pb.UrlResponse, error) {
	longUrl, err := h.urlService.GetLongUrl(ctx, data.Short)
	if err != nil {
		return nil, err
	}

	return &pb.UrlResponse{
		Short: "",
		Long:  longUrl,
		Error: "",
	}, nil
}
func (h *UrlHandler) SetShortUrl(ctx context.Context, data *pb.UrlRequest) (*pb.UrlResponse, error) {
	longUrl := data.Long
	shortUrl, err := h.urlService.SetShortUrl(ctx, longUrl)
	if err != nil {
		return nil, err
	}

	return &pb.UrlResponse{
		Short: shortUrl,
		Long:  longUrl,
		Error: "",
	}, nil
}
