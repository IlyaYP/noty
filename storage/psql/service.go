package psql

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/rs/zerolog"
	"noty/pkg/logging"
	"noty/storage"
)

var _ storage.Storage = (*Storage)(nil)

//var _ storage.OrderStorage = (*Storage)(nil)

const (
	serviceName = "psql"
)

type (
	Storage struct {
		config Config
		pool   *pgxpool.Pool
		ctx    context.Context
	}

	option func(svc *Storage) error
)

// WithConfig sets Config.
func WithConfig(cfg Config) option {
	return func(svc *Storage) error {
		svc.config = cfg
		return nil
	}
}

// WithContext sets Context.
func WithContext(ctx context.Context) option {
	return func(svc *Storage) error {
		svc.ctx = ctx
		return nil
	}
}

// New creates a new Storage.
func New(opts ...option) (*Storage, error) {
	svc := &Storage{
		config: NewDefaultConfig(),
	}

	for _, opt := range opts {
		if err := opt(svc); err != nil {
			return nil, fmt.Errorf("initialising dependencies: %w", err)
		}
	}

	if err := svc.config.validate(); err != nil {
		return nil, fmt.Errorf("Config validation: %w", err)
	}

	if svc.ctx == nil {
		svc.ctx = context.Background()
	}

	pool, err := pgxpool.Connect(svc.ctx, svc.config.DSN)
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %w", err)
	}
	svc.pool = pool

	if err := svc.Ping(svc.ctx); err != nil {
		return nil, fmt.Errorf("ping for DSN (%s) failed: %w", svc.config.DSN, err)
	}

	if err := svc.Migrate(svc.ctx); err != nil {
		return nil, fmt.Errorf("unable to create table: %w", err)
	}

	return svc, nil
}

func (svc *Storage) Migrate(ctx context.Context) error {
	logger := svc.Logger(ctx)
	logger.Info().Msg("Creating Tables")

	// TODO:for accrual int may be use money
	_, err := svc.pool.Exec(ctx, `
	DO
	$$
	BEGIN
		create extension if not exists "uuid-ossp";
		CREATE TABLE IF NOT EXISTS clients
		(
		    id uuid default uuid_generate_v4(),
			phone bigint not null,
			op_code int not null,
			tag varchar(64),
			tz varchar(64),
			primary key (id),
			unique (phone)
		);

		IF NOT EXISTS (SELECT *
		FROM pg_type typ
		INNER JOIN pg_namespace nsp
		ON nsp.oid = typ.typnamespace
		WHERE nsp.nspname = current_schema()
		AND typ.typname = 'filter') THEN
		CREATE TYPE filter AS (
		tags            text[],
		codes     bigint[]
		);
		END IF;
		
		CREATE TABLE IF NOT EXISTS sendings
		(
			id uuid default uuid_generate_v4(),
			start_at timestamp with time zone not null default now(),
			text varchar(160) not null,
			filter filter,
			stop_at timestamp with time zone not null default now(),
			primary key (id)
		);

		CREATE TABLE IF NOT EXISTS tags
		(
			sending_id uuid not null,
			tag varchar(64) not null,
			foreign key (sending_id) references sendings (id) ON DELETE CASCADE
		);
		
		CREATE TABLE IF NOT EXISTS codes
		(
			sending_id uuid not null,
			op_code int not null,
			foreign key (sending_id) references sendings (id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS messages
		(
			id bigserial not null,
			created_at timestamp with time zone not null default now(),
			status int not null,
		    sending_id uuid not null,
		    client_id uuid not null,
			primary key (id),
			foreign key (sending_id) references sendings (id) ON DELETE CASCADE,
			foreign key (client_id) references clients (id) ON DELETE CASCADE,
			unique (sending_id, client_id)
		);
	END;
	$$
	`)

	return err
}

func (svc *Storage) Destroy(ctx context.Context) error {
	logger := svc.Logger(ctx)
	logger.Info().Msg("Drop Tables")

	_, err := svc.pool.Exec(ctx, `drop table clients,sendings,messages;`)

	return err
}

// Ping checks db connection
func (svc *Storage) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, svc.config.timeout)
	defer cancel()

	return svc.pool.Ping(ctx)
}

// Close closes DB connection.
func (svc *Storage) Close() error {
	logger := svc.Logger(svc.ctx)
	logger.Info().Msg("Closing storage")
	if svc.pool != nil {
		svc.pool.Close()
	}
	return nil
}

// Logger returns logger with Storage field set.
func (svc *Storage) Logger(ctx context.Context) *zerolog.Logger {
	_, logger := logging.GetCtxLogger(ctx)
	logger = logger.With().Str(logging.ServiceKey, serviceName).Logger()

	return &logger
}
