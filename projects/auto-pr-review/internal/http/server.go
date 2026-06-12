package http

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"auto-pr-review/internal/service"
)

//go:embed static
var staticFiles embed.FS

type GitHubLinkService interface {
	Register(ctx context.Context, rawURL string) (service.GitHubLink, error)
	List(ctx context.Context) ([]service.GitHubLink, error)
	Delete(ctx context.Context, id int64) error
}

type PeriodicJobService interface {
	Register(ctx context.Context, rawURL, rawPeriod string) (service.PeriodicJob, error)
}

type Server struct {
	githubLinks    GitHubLinkService
	periodicJobs   PeriodicJobService
	logger         *slog.Logger
	requestTimeout time.Duration
}

func NewServer(
	githubLinks GitHubLinkService,
	periodicJobs PeriodicJobService,
	logger *slog.Logger,
	requestTimeout time.Duration,
) http.Handler {
	s := &Server{
		githubLinks:    githubLinks,
		periodicJobs:   periodicJobs,
		logger:         logger,
		requestTimeout: requestTimeout,
	}

	mux := http.NewServeMux()
	mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(mustStaticAssets()))))
	mux.HandleFunc("GET /", s.index)
	mux.HandleFunc("GET /index.html", s.index)
	mux.HandleFunc("POST /post", s.registerGitHubLink)
	mux.HandleFunc("GET /post", s.listGitHubLinks)
	mux.HandleFunc("DELETE /post/{id}", s.deleteGitHubLink)
	mux.HandleFunc("POST /jobs", s.registerPeriodicJob)
	mux.HandleFunc("GET /healthz", s.healthz)

	return s.withRequestTimeout(mux)
}

func (s *Server) index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		http.NotFound(w, r)
		return
	}
	page, err := staticFiles.ReadFile("static/index.html")
	if err != nil {
		s.logger.Error("read index page", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(page)
}

func (s *Server) registerGitHubLink(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	link, err := s.githubLinks.Register(r.Context(), req.URL)
	if err != nil {
		if errors.Is(err, service.ErrInvalidGitHubURL) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrGitHubLinkAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		s.logger.Error("register github link", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, link)
}

func (s *Server) registerPeriodicJob(w http.ResponseWriter, r *http.Request) {
	var req struct {
		URL    string `json:"url"`
		Period string `json:"period"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	job, err := s.periodicJobs.Register(r.Context(), req.URL, req.Period)
	if err != nil {
		if errors.Is(err, service.ErrInvalidGitHubURL) || errors.Is(err, service.ErrInvalidJobPeriod) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		s.logger.Error("register periodic job", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, job)
}

func (s *Server) listGitHubLinks(w http.ResponseWriter, r *http.Request) {
	links, err := s.githubLinks.List(r.Context())
	if err != nil {
		s.logger.Error("list github links", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, map[string][]service.GitHubLink{"links": links})
}

func (s *Server) deleteGitHubLink(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, service.ErrInvalidGitHubLinkID.Error())
		return
	}

	err = s.githubLinks.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidGitHubLinkID) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, service.ErrGitHubLinkNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		s.logger.Error("delete github link", slog.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) healthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) withRequestTimeout(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), s.requestTimeout)
		defer cancel()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func writeJSON(w http.ResponseWriter, statusCode int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, map[string]string{"error": message})
}

func mustStaticAssets() fs.FS {
	assets, err := fs.Sub(staticFiles, "static/assets")
	if err != nil {
		panic(err)
	}
	return assets
}
