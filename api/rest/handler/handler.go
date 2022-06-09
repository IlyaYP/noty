package handler

import (
	"compress/flate"
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/rs/zerolog"
	"net/http"
	"noty/pkg/logging"
	"noty/sender"
	"noty/storage"
)

const (
	serviceName = "handler"
)

type (
	Handler struct {
		*chi.Mux
		//tokenAuth *jwtauth.JWTAuth
		st  storage.Storage
		snd sender.Service
	}
	Option func(h *Handler) error
)

func NewHandler(opts ...Option) (*Handler, error) {
	h := &Handler{
		Mux: chi.NewMux(),
	}

	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, fmt.Errorf("initialising dependencies: %w", err)
		}
	}

	if h.st == nil {
		return nil, fmt.Errorf("st: nil")
	}

	if h.snd == nil {
		return nil, fmt.Errorf("snd: nil")
	}

	//
	//if h.orderSvc == nil {
	//	return nil, fmt.Errorf("orderSvc: nil")
	//}

	//h.tokenAuth = jwtauth.New("HS256", []byte("GMartSuperSecret"), nil)

	h.Use(middleware.Logger)
	compressor := middleware.NewCompressor(flate.DefaultCompression)
	h.Use(compressor.Handler)
	h.Use(render.SetContentType(render.ContentTypeJSON))
	h.Use(middleware.Recoverer)

	h.MethodNotAllowed(methodNotAllowedHandler)
	h.NotFound(notFoundHandler)
	h.Route("/api/client", h.client)
	h.Route("/api/sending", h.sending)
	h.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	return h, nil
}

// WithStorage sets Storage.
func WithStorage(st storage.Storage) Option {
	return func(h *Handler) error {
		h.st = st
		return nil
	}
}

// WithSender sets Sender.
func WithSender(snd sender.Service) Option {
	return func(h *Handler) error {
		h.snd = snd
		return nil
	}
}

func (h *Handler) Logger(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().Str(logging.ServiceKey, serviceName).Logger()

	return &logger
}

func methodNotAllowedHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(405)
	render.Render(w, r, ErrMethodNotAllowed)
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(400)
	render.Render(w, r, ErrNotFound)
}
