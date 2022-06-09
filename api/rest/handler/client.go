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

func (h *Handler) client(router chi.Router) {
	//router.Get("/", h.clientsGet)
	router.Get("/", h.clientsFilter)
	router.Post("/", h.clientAdd)
	router.Route("/{id}", func(router chi.Router) {
		router.Use(h.clientContext)
		router.Put("/", h.clientUpdate)
		router.Delete("/", h.clientDelete)
	})
}

func (h *Handler) clientsFilter(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	filter := model.Filter{
		Tags:  []string{"vip1", "vip2"},
		Codes: []int{911, 912},
	}
	clients, err := h.st.FilterClients(ctx, filter)
	if err != nil {
		if errors.Is(err, pkg.ErrNoData) {
			render.Render(w, r, ErrNoData)
			return
		}
		logger.Err(err).Msg("clientsGet: can't get clients from DB")
		render.Render(w, r, ErrServerError(err))
		return
	}

	render.Render(w, r, clients)

}

func (h *Handler) clientsGet(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	clients, err := h.st.GetClients(ctx)
	if err != nil {
		if errors.Is(err, pkg.ErrNoData) {
			render.Render(w, r, ErrNoData)
			return
		}
		logger.Err(err).Msg("clientsGet: can't get clients from DB")
		render.Render(w, r, ErrServerError(err))
		return
	}

	render.Render(w, r, clients)

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

	if input.ID == uuid.Nil {
		input.ID, _ = uuid.NewUUID()
	}

	logger.UpdateContext(input.GetLoggerContext)

	logger.Debug().Msg("This message appears only when log level set to Debug")

	_, err := h.st.CreateClient(ctx, *input)
	if err != nil {
		logger.Err(err).Msg("clientAdd st.CreateClient")
		if errors.Is(err, pkg.ErrAlreadyExists) {
			render.Render(w, r, ErrAlreadyExists)
			return
		}
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msg("new client")
	//render.Render(w, r, &client)
}

// clientContext do smth
func (h *Handler) clientContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		uid, err := uuid.Parse(id)
		if err != nil {
			render.Render(w, r, ErrInvalidRequest(err))
			return
		}
		ctx := context.WithValue(r.Context(), logging.ClientIDKey, uid)
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

	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	input.ID = uid

	logger.UpdateContext(input.GetLoggerContext)

	client, err := h.st.UpdateClient(ctx, *input)
	if err != nil {
		logger.Err(err).Msg("clientUpdate st.UpdateClient")
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msg("update client")

	render.Render(w, r, &client)

}

// clientDelete deletes client
// DELETE /api/client/{id}
func (h *Handler) clientDelete(w http.ResponseWriter, r *http.Request) {
	ctx, _ := logging.GetCtxLogger(r.Context())
	logger := h.Logger(ctx)

	id := chi.URLParam(r, "id")
	uid, err := uuid.Parse(id)
	if err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	if err := h.st.DeleteClientByID(ctx, uid); err != nil {
		render.Render(w, r, ErrInvalidRequest(err))
		return
	}

	logger.Info().Msgf("Delete client %s", uid)
	fmt.Fprintf(w, "Delete client %s", uid)

}
