package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/msyamsula/portofolio/backend-app/url-shortener/services"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

type handler struct {
	svc services.Service
}

type handlerResponse struct {
	Error    string `json:"error,omitempty"`
	ShortUrl string `json:"short_url,omitempty"`
}

func parseUri(c context.Context, uri string) error {
	var err error
	_, span := otel.Tracer("handler").Start(c, "parse request uri")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	_, err = url.ParseRequestURI(uri)
	return err
}

func (h *handler) Short(w http.ResponseWriter, req *http.Request) {
	var err error
	ctx, span := otel.Tracer("handler").Start(req.Context(), "Short")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	query := req.URL.Query()
	longUrl := query.Get("long_url")
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

	if err = parseUri(ctx, longUrl); err != nil {
		return
	}

	shortUrl, err = h.svc.Short(ctx, longUrl)
	if err != nil {
		return
	}
}

func (h *handler) Redirect(w http.ResponseWriter, req *http.Request) {
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.RedirectShortUrl")
	var err error
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}

		span.End()
	}()

	// url path to this block is in this format = /api/url/redirect/{key}
	paths := mux.Vars(req)
	var shortUrl string
	var ok bool
	if shortUrl, ok = paths["shortUrl"]; !ok {
		err = errors.New("bad short url")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var longUrl string
	longUrl, err = h.svc.GetLongUrl(ctx, shortUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	span.End() // end span before redirect

	// redirect to the longUrl
	http.Redirect(w, req, longUrl, http.StatusPermanentRedirect)
}
