package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"net/http"
	"noty/model"
	"noty/pkg"
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

	sendings, err := h.st.GetSendingsStatus(ctx)
	if err != nil {
		if errors.Is(err, pkg.ErrNoData) {
			render.Render(w, r, ErrNoData)
			return
		}
		logger.Err(err).Msg("sendingsGenStat: can't get sendings from DB")
		render.Render(w, r, ErrServerError(err))
		return
	}

	render.Render(w, r, &sendings)
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
	//logger.Info().Msgf("sendingStat %s", id)
	//fmt.Fprintf(w, "sendingStat  %s", id)

	uid, err := uuid.Parse(id)
	if err != nil {
		logger.Err(err).Msg("sendingStat uuid.Parse")
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	messages, err := h.st.GetMessagesBySendingID(ctx, uid)
	if err != nil {
		logger.Err(err).Msg("sendingStat GetMessagesBySendingID")
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	render.Render(w, r, &messages)
}

// sendingAdd adds new sending
func (h *Handler) sendingAdd(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context()) // I don't want to use context from r.Context()
	// because it is canceled after the request is done
	//ctx, _ := logging.GetCtxLogger(context.Background())

	logger := h.Logger(ctx)

	input := &model.Sending{}

	if err := render.Bind(r, input); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		logger.Err(err).Msg("sendingAdd render.Bind")
		return
	}

	if input.ID == uuid.Nil {
		input.ID, _ = uuid.NewUUID()
	}

	logger.UpdateContext(input.GetLoggerContext)
	ctx = logging.SetCtxLogger(ctx, *logger)

	_, err := h.st.CreateSending(ctx, *input)
	if err != nil {
		logger.Err(err).Msg("sendingAdd")
		if errors.Is(err, pkg.ErrAlreadyExists) {
			render.Render(w, r, ErrAlreadyExists)
			return
		}
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msg("new sending")
	logger.Debug().Msgf("sending: %+v", input)

	go h.snd.NewSending(ctx, *input)

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

	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	input.ID = uid

	logger.UpdateContext(input.GetLoggerContext)
	ctx = logging.SetCtxLogger(ctx, *logger)

	sending, err := h.st.UpdateSending(ctx, *input)
	if err != nil {
		logger.Err(err).Msg("sendingUpdate st.UpdateSending")
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msg("update sending")

	render.Render(w, r, &sending)
}

// sendingDelete deletes sending
// DELETE /sending/{id}/
func (h *Handler) sendingDelete(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := h.st.DeleteSendingByID(ctx, uid); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msgf("Delete sending %s", id)
	fmt.Fprintf(w, "Delete sending %s", id)
}
