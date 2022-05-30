package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"noty/api/rest"
	"noty/pkg/logging"
	"noty/storage/psql"
	"os"
	"os/signal"
	"syscall"
)

const (
	LogLevel = zerolog.TraceLevel
)

func main() {
	log.Logger = log.Logger.
		Output(zerolog.ConsoleWriter{
			Out: os.Stderr,
			//TimeFormat: time.RFC3339,
		}).
		Level(LogLevel)

	if err := run(); err != nil {
		log.Fatal().Err(err).Msg("can't start app")
	}
	os.Exit(0)
}

func run() error {
	ctx, logger := logging.GetCtxLogger(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	st, err := psql.New(
		//psql.WithConfig(c.PSQLStorage),
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

	srv, err := rest.New(rest.WithStorage(st))
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
