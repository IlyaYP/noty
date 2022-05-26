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
		CREATE TABLE IF NOT EXISTS sendings
		(
		    id uuid default uuid_generate_v4(),
			start_at timestamp not null default now(),
			text varchar(160) not null,
			filter varchar(100),
			stop_at timestamp,
			primary key (id)
		);
		CREATE TABLE IF NOT EXISTS messages
		(
			id bigserial not null,
			created_at timestamp not null default now(),
			status int not null,
		    sending_id uuid not null,
		    client_id uuid not null,
			primary key (id),
			foreign key (sending_id) references sendings (id),
			foreign key (client_id) references clients (id)
		);
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
