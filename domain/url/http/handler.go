package http

import (
	"encoding/json"
	"net/http"
	neturl "net/url"

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

	type response struct {
		Error    string `json:"error"`
		ShortUrl string `json:"short_url,omitempty"`
	}
	resp := response{}
	var err error
	if err = checkUrl(longUrl); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	shortUrl, err := h.urlService.SetShortUrl(ctx, longUrl)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		resp.Error = err.Error()
		json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp.ShortUrl = shortUrl
	json.NewEncoder(w).Encode(resp)

}

func checkUrl(u string) error {
	if _, err := neturl.ParseRequestURI(u); err != nil {
		return err
	}

	return nil
}
