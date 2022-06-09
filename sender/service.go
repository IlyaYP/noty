package sender

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog"
	"noty/model"
	"noty/pkg/logging"
	"noty/storage"
	"time"
)

var _ Service = (*service)(nil)

const (
	serviceName = "sender-service"
)

type (
	service struct {
		Storage storage.Storage
		//AccrualProvider accrual.Provider
	}

	Option func(svc *service) error
)

//// WithAccrualProvider sets accrual.Provider.
//func WithAccrualProvider(pr accrual.Provider) Option {
//	return func(svc *service) error {
//		svc.AccrualProvider = pr
//		return nil
//	}
//}

// WithStorage sets storage.Storage.
func WithStorage(st storage.Storage) Option {
	return func(svc *service) error {
		svc.Storage = st
		return nil
	}
}

// New creates a new service.
func New(opts ...Option) (*service, error) {
	svc := &service{}
	for _, opt := range opts {
		if err := opt(svc); err != nil {
			return nil, fmt.Errorf("initialising dependencies: %w", err)
		}
	}

	if svc.Storage == nil {
		return nil, fmt.Errorf("storage: nil")
	}
	//
	//if svc.AccrualProvider == nil {
	//	return nil, fmt.Errorf("AccrualProvider: nil")
	//}
	return svc, nil

}

func (svc *service) Process(ctx context.Context, sending model.Sending) error {
	//ctx, _ = logging.GetCtxLogger(ctx) // correlationID is created here
	logger := svc.Logger(ctx)

	logger.Info().Msgf("Process...")

	if !svc.CheckTime(ctx, sending) {
		return nil
	}

	clients, err := svc.Storage.FilterClients(ctx, sending.Filter)
	if err != nil {
		logger.Err(err).Msg("failed to filter clients")
		return fmt.Errorf("filtering clients: %w", err)
	}

	cl, _ := json.Marshal(clients)
	logger.Debug().Msgf("clients: %s", string(cl))

	for _, client := range clients {
		message, err := svc.Storage.CreateMessage(ctx,
			model.Message{
				Status:    model.MessageStatusNew,
				SendingID: sending.ID,
				ClientID:  client.ID,
			})
		if err != nil {
			logger.Err(err).Msg("failed to create message")
			continue
		}
		logger.Debug().Msgf("client: %v message: %v", client.Phone, message.ID)
	}

	return nil
}

// Logger returns logger with service field set.
func (svc *service) Logger(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().Str(logging.ServiceKey, serviceName).Logger()

	return &logger
}

// CheckTime checks if time is between start and end.
func (svc *service) CheckTime(ctx context.Context, sending model.Sending) bool {
	if sending.StartAt.Before(time.Now()) && sending.StopAt.After(time.Now()) {
		return true
	}
	return false
}
