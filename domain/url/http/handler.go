package http

import (
	"encoding/json"
	"net/http"

	url "github.com/msyamsula/portofolio/domain/url/service"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Handler struct {
	urlService *url.Service
	tracer     trace.Tracer
}

type Dependencies struct {
	UrlService *url.Service
}

func New(dep Dependencies) *Handler {
	return &Handler{
		urlService: dep.UrlService,
	}
}

func (h *Handler) HashUrl(w http.ResponseWriter, req *http.Request) {

	HashCounter.Inc()
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.HashUrl")
	defer span.End()

	query := req.URL.Query()
	longUrl := query.Get("long_url")
	shortUrl, err := h.urlService.SetShortUrl(ctx, longUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
	}

	response := struct {
		ShortUrl string `json:"short_url,omitempty"`
	}{
		ShortUrl: shortUrl,
	}
	resp, err := json.Marshal(response)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(resp)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}
