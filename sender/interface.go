package sender

import (
	"context"
	"noty/model"
)

type Service interface {
	ProcessSending(ctx context.Context, sending model.Sending) error

	ProcessSendings(ctx context.Context) error
}
