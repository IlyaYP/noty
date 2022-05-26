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

// TODO: find out about: - обработки активных рассылок и отправки сообщений клиентам

func (h *Handler) sending(router chi.Router) {
	router.Get("/", h.sendingsGenStat)
	router.Post("/", h.sendingAdd)
	router.Route("/{id}", func(router chi.Router) {
		router.Use(h.sendingContext)
		router.Get("/", h.sendingStat)
		router.Put("/", h.sendingUpdate)
		router.Delete("/", h.sendingDelete)
	})
}

// sendingsGenStat
// obtaining general statistics on created sendings and the number of sent
// messages on them, grouped by status
func (h *Handler) sendingsGenStat(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	logger.Info().Msgf("sendingsGenStat")
	fmt.Fprintf(w, "sendingsGenStat")
}

// sendingContext do smth
func (h *Handler) sendingContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		ctx := context.WithValue(r.Context(), logging.SendingIDKey, id)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sendingStat
// obtaining detailed statistics of sent messages for a specific Sending
func (h *Handler) sendingStat(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	id := chi.URLParam(r, "id")
	logger.Info().Msgf("sendingStat %s", id)
	fmt.Fprintf(w, "sendingStat  %s", id)
}

// sendingAdd adds new sending
func (h *Handler) sendingAdd(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	input := &model.Sending{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		logger.Err(err).Msg("sendingAdd render.Bind")
		return
	}

	input.ID, _ = uuid.NewUUID()
	logger.UpdateContext(input.GetLoggerContext)

	logger.Info().Msg("new sending")

	render.Render(w, r, input)

}

// sendingUpdate updates sending
func (h *Handler) sendingUpdate(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	input := &model.Sending{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		logger.Err(err).Msg("sendingUpdate render.Bind")
		return
	}

	logger.UpdateContext(input.GetLoggerContext)

	//user, err := h.userSvc.Register(ctx, input.Login, input.Password)

	logger.Info().Msg("update sending")

	render.Render(w, r, input)
}

// sendingDelete deletes sending
// DELETE /sending/{id}/
func (h *Handler) sendingDelete(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	id := chi.URLParam(r, "id")
	logger.Info().Msgf("Delete sending %s", id)
	fmt.Fprintf(w, "Delete sending %s", id)
}
