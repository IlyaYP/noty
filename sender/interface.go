package sender

import (
	"context"
	"noty/model"
)

type Service interface {
	NewSending(ctx context.Context, sending model.Sending)
}
