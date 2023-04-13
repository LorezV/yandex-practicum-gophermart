package main

import (
	"context"
	"github.com/LorezV/go-diploma.git/internal/accural"
	"github.com/LorezV/go-diploma.git/internal/config"
	"github.com/LorezV/go-diploma.git/internal/database"
	"github.com/LorezV/go-diploma.git/internal/repository"
	"github.com/LorezV/go-diploma.git/internal/server"
	"github.com/LorezV/go-diploma.git/internal/services"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "03:04:05PM"})

	cfg := config.MakeConfig()

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.MakeConnection(mainCtx, cfg.DatabaseURI)
	if err != nil {
		log.Error().Err(err).Send()
		return
	}
	defer db.Close()

	accrualClient := clients.MakeAccrualClient(cfg.AccrualSystemAddress)
	mainRepository := repository.MakeRepository(db)
	mainServices := services.MakeServices(mainRepository, accrualClient, config.SecretKey)

	s := server.NewServer(cfg.RunAddress, mainServices)

	g, gCtx := errgroup.WithContext(mainCtx)
	g.Go(func() error {
		return s.Run()
	})
	g.Go(func() error {
		<-gCtx.Done()
		return s.Shutdown(context.Background())
	})
	g.Go(func() error {
		if err = mainServices.Order.RunPolling(mainCtx); err != nil {
			log.Error().Err(err).Msg("Failed polling statuses")
			return err
		}
		return nil
	})

	if err = g.Wait(); err != nil {
		log.Info().Msg("The application is shutdown")
	}
}
