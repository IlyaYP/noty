package psql

import (
	"context"
	"errors"
	"github.com/jackc/pgconn"
	"noty/model"
	"noty/pkg"
)

func (svc *Storage) CreateClient(ctx context.Context, client model.Client) (model.Client, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(client.GetLoggerContext)

	_, err := svc.pool.Exec(ctx,
		`insert into clients(phone, op_code, tag, tz) values ($1, $2, $3, $4, $5);`,
		client.ID, client.Phone, client.OpCode, client.Tag, client.TZ)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Err(err).Msg("Error creating sending")
				return model.Client{}, pkg.ErrAlreadyExists
			}
		}

		logger.Err(err).Msg("Error creating sending")
		return model.Client{}, err
	}

	logger.Info().Msg("Successfully created sending")

	return client, nil
}
