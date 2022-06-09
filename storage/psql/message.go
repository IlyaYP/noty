package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"noty/model"
)

func (svc *Storage) CreateMessage(ctx context.Context, message model.Message) (model.Message, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(message.GetLoggerContext)

	Row := svc.pool.QueryRow(ctx,
		`insert into messages(status, sending_id, client_id) values ($1, $2, $3) returning id`,
		message.Status.Int(), message.SendingID, message.ClientID)

	err := Row.Scan(&message.ID)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger.Err(pgErr).Msg("creating message")
			return model.Message{}, pgErr
		}

		logger.Err(err).Msg("creating message")
		return model.Message{}, err
	}

	logger.Info().Msg("Successfully created message")

	return message, nil
}
