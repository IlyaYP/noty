package sender

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
	"math/rand"
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

	rand.Seed(time.Now().UnixNano())

	return svc, nil

}

func (svc *service) ProcessSending(ctx context.Context, sending model.Sending) error {
	//ctx, _ = logging.GetCtxLogger(ctx) // correlationID is created here
	logger := svc.Logger(ctx)

	//logger.Info().Msgf("ProcessSending...")

	if !svc.CheckTime(ctx, sending) {
		return nil
	}

	clients, err := svc.Storage.FilterClients(ctx, sending.Filter)
	if err != nil {
		logger.Err(err).Msg("failed to filter clients")
		return fmt.Errorf("filtering clients: %w", err)
	}

	//cl, _ := json.Marshal(clients)
	//logger.Debug().Msgf("clients: %s", string(cl))

	for _, client := range clients {
		message, err := svc.Storage.GetMessageByClientAndSendingID(ctx, client.ID, sending.ID)
		if err == pgx.ErrNoRows {
			message, err = svc.Storage.CreateMessage(ctx,
				model.Message{
					Status:    model.MessageStatusNew,
					SendingID: sending.ID,
					ClientID:  client.ID,
				})
			if err != nil {
				logger.Err(err).Msg("failed to create message")
				continue
			}
		} else if err != nil {
			logger.Err(err).Msg("failed to get message")
			continue
		}

		if message.Status != model.MessageStatusNew {
			continue
		}

		err = svc.SendMessage(ctx, model.MessageToSend{
			ID:    message.ID,
			Phone: client.Phone,
			Text:  sending.Text,
		})

		if err != nil {
			logger.Err(err).Msgf("failed to send message: %v", message.ID)
			continue
		}

		message.Status = model.MessageStatusSent
		message.CreatedAt = time.Now()
		message, err = svc.Storage.UpdateMessage(ctx, message)
		if err != nil {
			logger.Err(err).Msgf("failed to update message: %v", message.ID)
			continue
		}

		logger.Debug().Msgf("client: %+v", client)
		logger.Debug().Msgf("message: %+v", message)
	}

	return nil
}

func (svc *service) ProcessSendings(ctx context.Context) error {
	logger := svc.Logger(ctx)
	logger.Info().Msg("started")

	sendings, err := svc.Storage.FilterCurrentSendings(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to filter sendings")
		return fmt.Errorf("filtering sendings: %w", err)
	}

	for _, sending := range sendings {
		err = svc.ProcessSending(ctx, sending)
		if err != nil {
			logger.Err(err).Msg("failed to process sending")
			continue
		}

		logger.Debug().Msgf("sending: %+v", sending)
	}

	logger.Info().Msg("finished")
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

// SendMessage sends message to client.
func (svc *service) SendMessage(ctx context.Context, message model.MessageToSend) error {
	logger := svc.Logger(ctx)

	sendingDelay := 3

	delay := rand.Intn(sendingDelay)
	if delay == 2 {
		return fmt.Errorf("cant send message")
	}

	logger.Info().Msgf("Sending message: %+v", message)
	time.Sleep(time.Duration(delay) * time.Second)

	return nil
}

// Run starts service.
func (svc *service) Run(ctx context.Context) error {
	logger := svc.Logger(ctx)
	logger.Info().Msg("started")

	for {
		select {
		case <-ctx.Done():
			logger.Info().Msg("stopped")
			return nil
		case <-time.After(time.Second * 10):
			err := svc.ProcessSendings(ctx)
			if err != nil {
				logger.Err(err).Msg("failed to process sendings")
			}
		}
	}
}
