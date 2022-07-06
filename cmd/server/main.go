package main

import (
	"context"
	"log"
	"os"

	"go.uber.org/zap"

	"github.com/jeffreyyong/news-feeder/internal/app"
	"github.com/jeffreyyong/news-feeder/internal/config"
	"github.com/jeffreyyong/news-feeder/internal/crawler"
	"github.com/jeffreyyong/news-feeder/internal/logging"
	"github.com/jeffreyyong/news-feeder/internal/rss"
	"github.com/jeffreyyong/news-feeder/internal/service"
	"github.com/jeffreyyong/news-feeder/internal/store"
	"github.com/jeffreyyong/news-feeder/pkg/apppostgres"
	"github.com/jmoiron/sqlx"
	cli "github.com/urfave/cli/v2"

	"github.com/pkg/errors"
)

var command string

func main() {
	app := &cli.App{
		Name:  "News Feeder",
		Usage: "Aggregates news and feed them",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "command",
				EnvVars:     []string{"COMMAND"},
				Usage:       "The command to run",
				Destination: &command,
			},
		},
		Action: func(ctx *cli.Context) error {
			if command == "server" {
				return ctx.App.Run([]string{ctx.App.Name, "s"})
			}
			if command == "worker" {
				return ctx.App.Run([]string{ctx.App.Name, "w"})
			}
			return ctx.App.Run([]string{ctx.App.Name, "h"})
		},
		Commands: []*cli.Command{
			workerCommand,
			serverCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func newStore(ctx context.Context, s *app.Service, cfg config.Config) (*store.Store, error) {
	// Postgres Store
	// The postgres driver comes with a liveness health probe, checks connectivity as part of client creation
	// and provides sensible timeout defaults along with setting up integration with Datadog, Tempo and Prometheus out of the box.
	// to understand the functionality and the options available to override default behaviour.
	postgresDB, err := apppostgres.NewClient(ctx, s, apppostgres.WithDSN(cfg.PostgresDSN))
	if err != nil {
		return nil, errors.Wrap(err, "creating_postgres_client")
	}
	s.OnShutdown(func() {
		if err := postgresDB.Close(); err != nil {
			logging.Error(ctx, "failed to close postgres connection", zap.Error(err))
		}
	})

	return store.New(sqlx.NewDb(postgresDB, "postgres"))
}

func newService(ctx context.Context, s *app.Service, cfg config.Config, store *store.Store) (*service.Service, error) {
	parser := rss.NewParser()
	crawler := crawler.New(parser, cfg.URLSources)
	svc, err := service.New(store, crawler)
	if err != nil {
		return nil, err
	}
	return svc, nil
}
