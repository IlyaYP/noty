package config

import (
	"flag"
	"fmt"
	"github.com/caarlos0/env/v6"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"noty/api/rest"
	"noty/sender"
	"noty/storage/psql"
)

// Config combines sub-configs for all services, storages and providers.
type Config struct {
	Sender        sender.Config
	PSQLStorage   psql.Config
	APISever      rest.Config
	Address       string `env:"RUN_ADDRESS"`
	SenderAddress string `env:"SENDER_ADDRESS"`
	SenderToken   string `env:"SENDER_TOKEN"`
	DSN           string `env:"DATABASE_URI"`
	Closer        []io.Closer
}

// New initializes a new config.
func New() (*Config, error) {
	cfg := Config{}
	flag.StringVar(&cfg.DSN, "d", "postgres://postgres:postgres@localhost:5432/noty", "DATABASE_URI")
	flag.StringVar(&cfg.Address, "a", "localhost:8080", "RUN_ADDRESS")
	flag.StringVar(&cfg.SenderAddress, "r", "localhost:8081", "SENDER_ADDRESS")
	flag.StringVar(&cfg.SenderToken, "t", "", "SENDER_TOKEN")
	debug := flag.Bool("debug", false, "sets log level to debug")
	flag.Parse()

	// Default level for this example is info, unless debug flag is present
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("initializing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	cfg.PSQLStorage = psql.NewDefaultConfig()
	cfg.PSQLStorage.DSN = cfg.DSN
	cfg.APISever.Address = cfg.Address
	cfg.Sender = sender.NewDefaultConfig()
	cfg.Sender.Address = cfg.SenderAddress
	cfg.Sender.Token = cfg.SenderToken

	return &cfg, nil
}

// validate performs a basic validation.
func (c Config) validate() error {
	if c.Address == "" {
		return fmt.Errorf("%s field: empty", "RUN_ADDRESS")
	}
	Logger := log.With().Str("RUN_ADDRESS", c.Address).Logger()

	if c.DSN == "" {
		return fmt.Errorf("%s field: empty", "DATABASE_URI")
	}
	Logger = Logger.With().Str("DATABASE_URI", c.DSN).Logger()

	if c.SenderAddress == "" {
		return fmt.Errorf("%s field: empty", "SENDER_ADDRESS")
	}
	Logger = Logger.With().Str("SENDER_ADDRESS", c.SenderAddress).Logger()

	Logger.Debug().Msg("Initialized with args:")

	return nil
}

//// BuildPsqlStorage builds psql.Storage dependency.
//func (c Config) BuildPsqlStorage(ctx context.Context) (*psql.Storage, error) {
//	st, err := psql.New(
//		psql.WithConfig(c.PSQLStorage),
//		psql.WithContext(ctx),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building psql storage: %w", err)
//	}
//
//	return st, nil
//}

//// BuildUserService builds user.Service dependency.
//func (c Config) BuildUserService(ctx context.Context) (user.Service, error) {
//	st, err := c.BuildPsqlStorage(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	svc, err := user.New(
//		user.WithConfig(c.UserService),
//		user.WithUserStorage(st),
//	)
//
//	if err != nil {
//		return nil, fmt.Errorf("building user service: %w", err)
//	}
//
//	return svc, nil
//
//}

//// BuildOrderService builds order.Service dependency.
//func (c Config) BuildOrderService(ctx context.Context) (order.Service, error) {
//	st, err := c.BuildPsqlStorage(ctx)
//	if err != nil {
//		return nil, err
//	}
//
//	// BuildAccrualProvider builds Accrual Provider dependency
//	accPr, err := c.BuildAccrualProvider(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("accrual provider: %w", err)
//	}
//
//	svc, err := order.New(
//		order.WithOrderStorage(st),
//		order.WithAccrualProvider(accPr),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building test service: %w", err)
//	}
//
//	return svc, nil
//}

//// BuildServer builds REST API Server dependency.
//func (c Config) BuildServer(ctx context.Context) (*server.Server, error) {
//	//userSvc, err := c.BuildUserService(ctx)
//	//if err != nil {
//	//	return nil, fmt.Errorf("building server: %w", err)
//	//}
//	//
//	//orderSvc, err := c.BuildOrderService(ctx)
//	//if err != nil {
//	//	return nil, fmt.Errorf("building server: %w", err)
//	//}
//
//	// Build Storage
//	st, err := c.BuildPsqlStorage(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("building server: %w", err)
//	}
//	c.Closer = append(c.Closer, st)
//
//	// Build User Service
//	userSvc, err := user.New(
//		user.WithConfig(c.UserService),
//		user.WithUserStorage(st),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building user service: %w", err)
//	}
//
//	// BuildAccrualProvider builds Accrual Provider dependency
//	accPr, err := c.BuildAccrualProvider(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("accrual provider: %w", err)
//	}
//
//	// Build Order Service
//	orderSvc, err := order.New(
//		order.WithOrderStorage(st),
//		order.WithAccrualProvider(accPr),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building order service: %w", err)
//	}
//
//	// Build REST API Service
//	r, err := handler.NewHandler(
//		handler.WithUserService(userSvc),
//		handler.WithOrderService(orderSvc),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building server: %w", err)
//	}
//
//	s, err := server.New(
//		server.WithConfig(&c.APISever),
//		server.WithRouter(r),
//	)
//	if err != nil {
//		return nil, fmt.Errorf("building server: %w", err)
//	}
//	return s, nil
//}

//// BuildAccrualProvider builds Accrual Provider dependency
//func (c Config) BuildAccrualProvider(ctx context.Context) (accrual.Provider, error) {
//	svc, err := http.New(http.WithConfig(&c.AccrualHTTPProvider))
//	if err != nil {
//		return nil, fmt.Errorf("building provider: %w", err)
//	}
//
//	return svc, nil
//}
