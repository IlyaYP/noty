package psql

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"noty/model"
	"noty/pkg"
)

func (svc *Storage) CreateClient(ctx context.Context, client model.Client) (model.Client, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(client.GetLoggerContext)

	_, err := svc.pool.Exec(ctx,
		`insert into clients(id, phone, op_code, tag, tz) values ($1, $2, $3, $4, $5);`,
		client.ID, client.Phone, client.OpCode, client.Tag, client.TZ)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				logger.Err(err).Msg("Error creating client")
				return model.Client{}, pkg.ErrAlreadyExists
			}
		}

		logger.Err(err).Msg("Error creating client")
		return model.Client{}, err
	}

	logger.Info().Msg("Successfully created client")

	return client, nil
}

func (svc *Storage) DeleteClientByID(ctx context.Context, id uuid.UUID) error {
	logger := svc.Logger(ctx)

	res, err := svc.pool.Exec(ctx,
		`DELETE FROM clients WHERE id=$1`, id)
	if err != nil {
		logger.Err(err).Msg("DeleteClientByID")
		return err
	}

	if res.RowsAffected() == 0 {
		logger.Err(pkg.ErrNotExists).Msg("DeleteClientByID")
		return pkg.ErrNotExists
	}

	logger.Info().Msgf("Delete client %s, %v rows affected", id, res.RowsAffected())

	return nil
}

func (svc *Storage) UpdateClient(ctx context.Context, client model.Client) (model.Client, error) {
	logger := svc.Logger(ctx)

	res, err := svc.pool.Exec(ctx,
		`UPDATE clients SET phone=$1, op_code=$2, tag=$3, tz=$4 WHERE id=$5;`,
		client.Phone, client.OpCode, client.Tag, client.TZ, client.ID)
	if err != nil {
		logger.Err(err).Msg("UpdateClient")
		return model.Client{}, err
	}

	if res.RowsAffected() == 0 {
		logger.Err(pkg.ErrNotExists).Msg("UpdateClient")
		return model.Client{}, pkg.ErrNotExists
	}

	logger.Info().Msgf("Update client %s, %v rows affected", client.ID, res.RowsAffected())

	return client, nil
}

func (svc *Storage) GetClients(ctx context.Context) (model.Clients, error) {
	logger := svc.Logger(ctx)
	var clients model.Clients

	clientsRows, err := svc.pool.Query(
		ctx,
		"select * from clients ORDER BY phone ASC",
	)
	if err != nil {
		logger.Err(err).Msg("GetClients")
		return nil, err //pgx.ErrNoRows
	}
	defer clientsRows.Close()

	for clientsRows.Next() {
		client := model.Client{}
		err := clientsRows.Scan(
			&client.ID,
			&client.Phone,
			&client.OpCode,
			&client.Tag,
			&client.TZ,
		)
		if err != nil {
			logger.Err(err).Msg("GetClients")
			continue
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, pkg.ErrNoData
	}

	return clients, nil
}

func (svc *Storage) FilterClients(ctx context.Context, filter model.Filter) (model.Clients, error) {
	logger := svc.Logger(ctx)
	var clients model.Clients

	// select * from clients where op_code in (911,912) AND tag in ('vip1','vip2');
	clientsRows, err := svc.pool.Query(
		ctx,
		"select * from clients WHERE op_code = ANY($1::INT[]) AND tag = ANY($2::text[])",
		filter.Codes,
		filter.Tags,
	)
	if err != nil {
		logger.Err(err).Msg("GetClients")
		return nil, err //pgx.ErrNoRows
	}
	defer clientsRows.Close()

	for clientsRows.Next() {
		client := model.Client{}
		err := clientsRows.Scan(
			&client.ID,
			&client.Phone,
			&client.OpCode,
			&client.Tag,
			&client.TZ,
		)
		if err != nil {
			logger.Err(err).Msg("GetClients")
			continue
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil, pkg.ErrNoData
	}

	return clients, nil
}
