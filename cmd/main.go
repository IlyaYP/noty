package main

import (
	"context"
	"log"
	"noty/api/rest"
	"noty/pkg/logging"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

func run() error {
	ctx, logger := logging.GetCtxLogger(context.Background())
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	srv, err := rest.New()
	if err != nil {
		logger.Err(err).Msg("Can not create rest server")
		return err
	}

	go func() {
		<-ctx.Done()
		logger.Info().Msg("Stopping server")
		srv.Close(ctx)
	}()

	logger.Info().Msg("Starting server")
	return srv.Serve(ctx)

}
