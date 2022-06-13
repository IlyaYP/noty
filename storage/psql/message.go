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

func (svc *Storage) CreateMessage(ctx context.Context, message model.Message) (model.Message, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(message.GetLoggerContext)

	Row := svc.pool.QueryRow(ctx,
		`insert into messages(status, sending_id, client_id) values ($1, $2, $3) returning id, created_at`,
		message.Status.Int(), message.SendingID, message.ClientID)

	err := Row.Scan(&message.ID, &message.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			logger.Err(pgErr).Msg("creating message")
			return model.Message{}, pgErr
		}

		logger.Err(err).Msg("creating message")
		return model.Message{}, err
	}

	//logger.Debug().Msgf("Message: %+v", message)

	return message, nil
}

// UpdateMessage updates message.
func (svc *Storage) UpdateMessage(ctx context.Context, message model.Message) (model.Message, error) {
	logger := svc.Logger(ctx)
	logger.UpdateContext(message.GetLoggerContext)

	res, err := svc.pool.Exec(ctx,
		`update messages set status = $1, created_at = $2 where id = $3`,
		message.Status.Int(), message.CreatedAt, message.ID)
	if err != nil {
		logger.Err(err).Msg("updating message")
		return model.Message{}, err
	}

	if res.RowsAffected() == 0 {
		logger.Err(pkg.ErrNotExists).Msg("updating message")
		return model.Message{}, pkg.ErrNotExists
	}

	return message, nil
}

func (svc *Storage) GetMessageByClientAndSendingID(ctx context.Context, clientID uuid.UUID, sendingID uuid.UUID) (model.Message, error) {
	logger := svc.Logger(ctx)

	var message model.Message
	var status int
	err := svc.pool.QueryRow(ctx,
		`select id, status, created_at from messages where client_id = $1 and sending_id = $2`,
		clientID, sendingID).Scan(&message.ID, &status, &message.CreatedAt)
	if err != nil {
		if err != pgx.ErrNoRows {
			logger.Err(err).Msg("getting message")
		}
		return model.Message{}, err
	}
	message.Status = model.NewMessageStatusFromInt(status)

	return message, nil
}

// GetMessagesBySendingID returns messages by sending ID.
func (svc *Storage) GetMessagesBySendingID(ctx context.Context, sendingID uuid.UUID) (model.Messages, error) {
	logger := svc.Logger(ctx)

	rows, err := svc.pool.Query(ctx,
		`select id, created_at, status, sending_id, client_id from messages where sending_id = $1`,
		sendingID)
	if err != nil {
		logger.Err(err).Msg("getting messages")
		return nil, err
	}
	defer rows.Close()

	var messages model.Messages
	for rows.Next() {
		var message model.Message
		var status int
		err := rows.Scan(&message.ID, &message.CreatedAt, &status, &message.SendingID, &message.ClientID)
		if err != nil {
			logger.Err(err).Msg("getting messages")
			return nil, err
		}
		message.Status = model.NewMessageStatusFromInt(status)
		messages = append(messages, &message)

	}

	if err = rows.Err(); err != nil {
		logger.Err(err).Msg("getting messages")
		return nil, err
	}

	if len(messages) == 0 {
		logger.Err(pkg.ErrNoData).Msg("getting messages")
		return nil, pkg.ErrNoData
	}

	return messages, nil
}
