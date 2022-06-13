package storage

import (
	"context"
	"github.com/google/uuid"
	"io"
	"noty/model"
)

// Storage defines models operations.
type Storage interface {
	io.Closer

	// CreateClient creates a new model.Client.
	// Returns ErrAlreadyExists if client exists.
	CreateClient(ctx context.Context, client model.Client) (model.Client, error)

	UpdateClient(ctx context.Context, client model.Client) (model.Client, error)

	DeleteClientByID(ctx context.Context, id uuid.UUID) error

	GetClients(ctx context.Context) (model.Clients, error)

	FilterClients(ctx context.Context, filter model.Filter) (model.Clients, error)

	CreateSending(ctx context.Context, sending model.Sending) (model.Sending, error)

	UpdateSending(ctx context.Context, sending model.Sending) (model.Sending, error)

	DeleteSendingByID(ctx context.Context, id uuid.UUID) error

	GetSendings(ctx context.Context) (model.Sendings, error)

	GetSendingsStatus(ctx context.Context) (model.SendingsStatus, error)

	FilterCurrentSendings(ctx context.Context) (model.Sendings, error)

	CreateMessage(ctx context.Context, message model.Message) (model.Message, error)

	UpdateMessage(ctx context.Context, message model.Message) (model.Message, error)

	GetMessagesBySendingID(ctx context.Context, sendingID uuid.UUID) (model.Messages, error)

	GetMessageByClientAndSendingID(ctx context.Context, clientID uuid.UUID, sendingID uuid.UUID) (model.Message, error)
}
