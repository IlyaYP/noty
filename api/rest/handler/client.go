package handler

import (
	"context"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"noty/model"
	"noty/pkg/logging"
)

func (h *Handler) client(router chi.Router) {
	router.Post("/", h.clientAdd)
	router.Route("/{id}", func(router chi.Router) {
		router.Use(h.clientContext)
		router.Put("/", h.clientUpdate)
		router.Delete("/", h.clientDelete)
	})
}

// clientAdd adds new client
func (h *Handler) clientAdd(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	input := &model.Client{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		logger.Err(err).Msg("clientAdd render.Bind")
		return
	}

	input.ID, _ = uuid.NewUUID()
	logger.UpdateContext(input.GetLoggerContext)

	//user, err := h.userSvc.Register(ctx, input.Login, input.Password)

	logger.Info().Msg("new client")

	render.Render(w, r, input)

}

// clientContext do smth
func (h *Handler) clientContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := context.WithValue(r.Context(), logging.ClientIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// clientUpdate updates client
func (h *Handler) clientUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	input := &model.Client{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		logger.Err(err).Msg("clientUpdate render.Bind")
		return
	}

	logger.UpdateContext(input.GetLoggerContext)

	//user, err := h.userSvc.Register(ctx, input.Login, input.Password)

	logger.Info().Msg("update client")

	render.Render(w, r, input)

}

// clientDelete deletes client
// DELETE /api/client/{id}
func (h *Handler) clientDelete(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	id := chi.URLParam(r, "id")
	logger.Info().Msgf("Delete client %s", id)
	fmt.Fprintf(w, "Delete client %s", id)
}
