package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"noty/api/rest"
	"noty/cmd/config"
	"noty/pkg/logging"
	"noty/sender"
	"noty/storage/psql"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	LogLevel = zerolog.TraceLevel
)

func main() {
	log.Logger = log.
		Output(zerolog.ConsoleWriter{
			Out:        os.Stderr,
			TimeFormat: time.Stamp,
		}).
		Level(LogLevel)

	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("can't start app")
	}
	os.Exit(0)
}

func run() error {
	ctx, logger := logging.GetCtxLogger(context.Background())
	logger = logger.With().Int("ver", 1).Logger()
	ctx = logging.SetCtxLogger(ctx, logger)

	cfg, err := config.New()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	st, err := psql.New(
		psql.WithConfig(cfg.PSQLStorage),
		psql.WithContext(ctx),
	)

	if err != nil {
		logger.Err(err).Msg("building psql storage")
		return fmt.Errorf("building psql storage: %w", err)
	}
	defer st.Close()

	err = st.Ping(ctx)
	if err != nil {
		logger.Err(err).Msg("ping psql storage")
		return fmt.Errorf("ping psql storage: %w", err)
	}

	sndr, err := sender.New(sender.WithStorage(st))
	if err != nil {
		logger.Err(err).Msg("Can not create sender")
		return err
	}

	go sndr.Run(ctx)

	srv, err := rest.New(rest.WithConfig(cfg.APISever), rest.WithStorage(st), rest.WithSender(sndr))
	if err != nil {
		logger.Err(err).Msg("Can not create rest server")
		return err
	}

	go func() {
		<-ctx.Done()
		srv.Close(ctx)
	}()

	return srv.Serve(ctx)

}
