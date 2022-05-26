package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"noty/model"
	"noty/pkg"
)

// CreateSending creates a new model.Sending.
// Returns ErrAlreadyExists if user exists.
func (svc *Storage) CreateSending(ctx context.Context, sending model.Sending) (model.Sending, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(sending.GetLoggerContext)

	_, err := svc.pool.Exec(ctx, `insert into sendings(text, filter) values ($1, $2)`,
		sending.Text, sending.Filter)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Err(err).Msg("Error creating sending")
				return model.Sending{}, pkg.ErrAlreadyExists
			}
		}

		logger.Err(err).Msg("Error creating sending")
		return model.Sending{}, err
	}

	logger.Info().Msg("Successfully created sending")

	return sending, nil
}
