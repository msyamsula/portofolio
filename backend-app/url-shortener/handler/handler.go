package handler

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/services"
	"go.opentelemetry.io/otel"
)

type handler struct {
	svc services.Service
}

type handlerResponse struct {
	Error    string `json:"error,omitempty"`
	ShortUrl string `json:"short_url,omitempty"`
}

func (h *handler) Short(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.HashUrl")
	defer span.End()

	query := req.URL.Query()
	longUrl := query.Get("long_url")
	var err error
	var shortUrl string
	var resp handlerResponse

	defer func() {
		if err != nil {
			resp.Error = err.Error()
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		resp.ShortUrl = shortUrl
		json.NewEncoder(w).Encode(&resp)
	}()

	if _, err = url.ParseRequestURI(longUrl); err != nil {
		return
	}

	shortUrl, err = h.svc.Short(ctx, longUrl)
	if err != nil {
		return
	}
}

func (h *handler) Redirect(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.RedirectShortUrl")
	defer span.End()

	// url path to this block is in this format = /api/url/redirect/{key}
	paths := mux.Vars(req)
	if _, ok := paths["shortUrl"]; !ok {
		http.Error(w, "no short url given", http.StatusBadRequest)
		return
	}
	shortUrl := paths["shortUrl"] // get shortUrl

	longUrl, err := h.svc.GetLongUrl(ctx, shortUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// redirect to the longUrl
	http.Redirect(w, req, longUrl, http.StatusPermanentRedirect)
}
