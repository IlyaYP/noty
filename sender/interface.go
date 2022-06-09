package sender

import (
	"context"
	"noty/model"
)

type Service interface {
	Process(ctx context.Context, sending model.Sending) error
}
