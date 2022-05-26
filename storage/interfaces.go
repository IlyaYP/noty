package storage

import (
	"context"
	"io"
	"noty/model"
)

// Storage defines models operations.
type Storage interface {
	io.Closer

	// CreateClient creates a new model.Client.
	// Returns ErrAlreadyExists if client exists.
	CreateClient(ctx context.Context, client model.Client) (model.Client, error)

	CreateSending(ctx context.Context, sending model.Sending) (model.Sending, error)

	CreateMessage(ctx context.Context, message model.Message) (model.Message, error)
}
