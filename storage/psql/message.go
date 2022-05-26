package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"noty/model"
	"noty/pkg"
)

func (svc *Storage) CreateMessage(ctx context.Context, message model.Message) (model.Message, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(message.GetLoggerContext)

	_, err := svc.pool.Exec(ctx,
		`insert into messages(status, sending_id, client_id) values ($1, $2, $3)`,
		message.Status, message.SendingID, message.ClientID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Err(err).Msg("Error creating sending")
				return model.Message{}, pkg.ErrAlreadyExists
			}
		}

		logger.Err(err).Msg("Error creating sending")
		return model.Message{}, err
	}

	logger.Info().Msg("Successfully created sending")

	return message, nil
}
