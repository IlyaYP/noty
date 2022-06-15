package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/rs/zerolog"
	"io/ioutil"
	"math/rand"
	"net/http"
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
		inQueue chan model.Sending
		client  *http.Client
		config  Config
	}

	Option func(svc *service) error
)

// WithConfig sets config.
func WithConfig(cfg Config) Option {
	return func(svc *service) error {
		svc.config = cfg
		return nil
	}
}

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
	svc.inQueue = make(chan model.Sending, 100)

	svc.client = &http.Client{}

	return svc, nil

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
		case sending := <-svc.inQueue:
			err := svc.ProcessSending(ctx, sending)
			if err != nil {
				logger.Err(err).Msg("failed to process sending")
			}
		case <-time.After(time.Second * 60):
			err := svc.ProcessSendings(ctx)
			if err != nil {
				logger.Err(err).Msg("failed to process sendings")
			}
		}
	}
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

	sendings, err := svc.Storage.FilterCurrentSendings(ctx)
	if err != nil {
		logger.Err(err).Msg("failed to filter sendings")
		return fmt.Errorf("filtering sendings: %w", err)
	}

	for _, sending := range sendings {
		err = svc.ProcessSending(ctx, *sending)
		if err != nil {
			logger.Err(err).Msg("failed to process sending")
			continue
		}

		//logger.Debug().Msgf("sending: %+v", sending)
	}

	return nil
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

// Logger returns logger with service field set.
func (svc *service) Logger(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().Str(logging.ServiceKey, serviceName).Logger()

	return &logger
}

func (svc *service) NewSending(ctx context.Context, sending model.Sending) {
	svc.inQueue <- sending
}

// SendHTTP sends message to client.
// POST https://probe.fbrq.cloud/v1/send/{{msgId}}
func (svc *service) SendHTTP(ctx context.Context, message model.MessageToSend) error {
	logger := svc.Logger(ctx)

	jsonData, err := json.Marshal(message)
	if err != nil {
		logger.Err(err).Msg("failed to marshal message")
		return fmt.Errorf("marshaling message: %w", err)
	}

	url := fmt.Sprintf("%s/%d", svc.config.endPoint, message.ID)
	logger.Debug().Msgf("url: %s", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Err(err).Msg("failed to create request")
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", svc.config.Token))

	resp, err := svc.client.Do(req)
	if err != nil {
		logger.Err(err).Msg("failed to send request")
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Err(err).Msg("failed to send request")
		return fmt.Errorf("sending request: %w", err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Err(err).Msg("failed to read response")
		return fmt.Errorf("reading response: %w", err)
	}
	logger.Debug().Msgf("response: %s", string(body))

	return nil
}
