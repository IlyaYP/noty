package psql

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"noty/model"
	"noty/pkg"
)

// CreateSending creates a new model.Sending.
// Returns ErrAlreadyExists if user exists.
func (svc *Storage) CreateSending(ctx context.Context, sending model.Sending) (model.Sending, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(sending.GetLoggerContext)

	// insert into sendings(text, filter) values ('hello world!', ('{"vip1","vip2"}','{911, 912, 913}'));
	_, err := svc.pool.Exec(ctx,
		`insert into sendings(id, start_at, text, filter, stop_at) values ($1, $2, $3, $4, $5)`,
		sending.ID,
		sending.StartAt, sending.Text, sending.Filter, sending.StopAt)
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

func (svc *Storage) DeleteSendingByID(ctx context.Context, id uuid.UUID) error {
	logger := svc.Logger(ctx)

	res, err := svc.pool.Exec(ctx, `DELETE FROM sendings WHERE id=$1`, id)
	if err != nil {
		logger.Err(err).Msg("Error deleting sending")
		return err
	}

	logger.Info().Msgf("Delete sending %s, %v rows affected", id, res.RowsAffected())

	return nil
}

func (svc *Storage) UpdateSending(ctx context.Context, sending model.Sending) (model.Sending, error) {
	logger := svc.Logger(ctx)

	res, err := svc.pool.Exec(ctx,
		`UPDATE public.sendings SET start_at=$1, text=$2, filter=$3, stop_at=$4 WHERE id=$5;`,
		sending.StartAt, sending.Text, sending.Filter, sending.StopAt, sending.ID)
	if err != nil {
		logger.Err(err).Msg("UpdateSending")
		return model.Sending{}, err
	}

	if res.RowsAffected() == 0 {
		logger.Err(pkg.ErrNotExists).Msg("UpdateSending")
		return model.Sending{}, pkg.ErrNotExists
	}

	logger.Info().Msgf("Update sending %s, %v rows affected", sending.ID, res.RowsAffected())

	return sending, nil
}

func (svc *Storage) GetSendings(ctx context.Context) (model.Sendings, error) {
	logger := svc.Logger(ctx)
	var sendings model.Sendings

	sendingsRows, err := svc.pool.Query(
		ctx,
		"select * from sendings ORDER BY start_at ASC",
		pgx.QueryResultFormats{pgx.BinaryFormatCode},
	)
	if err != nil {
		logger.Err(err).Msg("GetSendings")
		return nil, err //pgx.ErrNoRows
	}
	defer sendingsRows.Close()

	for sendingsRows.Next() {
		sending := model.Sending{}
		err := sendingsRows.Scan(
			&sending.ID,
			&sending.StartAt,
			&sending.Text,
			&sending.Filter,
			&sending.StopAt,
		)
		if err != nil {
			logger.Err(err).Msg("GetSendings")
			continue
		}
		sendings = append(sendings, sending)
	}

	if len(sendings) == 0 {
		return nil, pkg.ErrNoData
	}

	return sendings, nil
}
