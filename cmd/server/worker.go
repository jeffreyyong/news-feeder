package main

import (
	"context"
	"time"

	"github.com/jeffreyyong/news-feeder/internal/app"
	"github.com/jeffreyyong/news-feeder/internal/app/listeners/worker"
	"github.com/jeffreyyong/news-feeder/internal/config"
	"github.com/jeffreyyong/news-feeder/internal/logging"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	workerServiceName = "worker"
)

var workerCommand = &cli.Command{
	Name:    "worker",
	Aliases: []string{"w"},
	Usage:   "Starts the worker process.",
	Action:  workerAction,
}

func workerAction(ctx *cli.Context) error {
	if err := app.Run(workerServiceName, workerSetup); err != nil {
		logging.From(ctx.Context).Fatal("failed to start worker",
			zap.String("service", workerServiceName),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func workerSetup(ctx context.Context, s *app.Service) ([]app.Listener, context.Context, error) {
	s.OnShutdown(func() {
		logging.Print(ctx, "shutdown",
			zap.String("service", workerServiceName),
		)
	})

	cfg, err := config.Load()
	if err != nil {
		return nil, ctx, errors.Wrap(err, "loading_config")
	}

	store, err := newStore(ctx, s, cfg)
	if err != nil {
		return nil, ctx, errors.Wrap(err, "unable to create store")
	}

	svc, err := newService(ctx, s, cfg, store)
	if err != nil {
		return nil, ctx, errors.Wrap(err, "unable to create service")
	}

	return []app.Listener{
		worker.New(svc, time.Duration(cfg.Worker.Interval)*time.Second),
	}, ctx, nil
}
