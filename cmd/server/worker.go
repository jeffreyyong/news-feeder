package main

import (
	"context"
	"os"
	"time"

	"github.com/jeffreyyong/news-feeder/internal/app"
	"github.com/jeffreyyong/news-feeder/internal/app/listeners/worker"
	"github.com/jeffreyyong/news-feeder/internal/config"
	"github.com/jeffreyyong/news-feeder/logging"
	"github.com/pkg/errors"
	cli "github.com/urfave/cli/v2"
	"go.uber.org/zap"
)

const (
	crawlerServiceName = "new-crawler"
)

var crawlerCommand = &cli.Command{
	Name:    "crawler",
	Aliases: []string{"c"},
	Usage:   "Starts the crawler process.",
	Action:  crawlerAction,
}

func crawlerAction(ctx *cli.Context) error {
	if err := app.Run(crawlerServiceName, crawlerSetup); err != nil {
		logging.From(ctx.Context).Fatal("failed to start crawler",
			zap.String("env", os.Getenv("CURVE_NAMESPACE")),
			zap.String("service", crawlerServiceName),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func crawlerSetup(ctx context.Context, s *app.Service) ([]app.Listener, context.Context, error) {
	s.OnShutdown(func() {
		logging.Print(ctx, "shutdown",
			zap.String("env", os.Getenv("CURVE_NAMESPACE")),
			zap.String("service", crawlerServiceName),
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
