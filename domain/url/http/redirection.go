package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/otel"
)

func (h *Handler) RedirectShortUrl(w http.ResponseWriter, req *http.Request) {
	RedirectCounter.Inc()
	ctx, span := otel.Tracer("").Start(req.Context(), "handler.RedirectShortUrl")
	defer span.End()

	// url path to this block is in this format = /api/url/redirect/{key}
	paths := mux.Vars(req)
	if _, ok := paths["shortUrl"]; !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	shortUrl := paths["shortUrl"] // get shortUrl

	// check the db
	longUrl, err := h.urlService.GetLongUrl(ctx, shortUrl)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	// redirect to the longUrl
	http.Redirect(w, req, longUrl, http.StatusSeeOther)

}
