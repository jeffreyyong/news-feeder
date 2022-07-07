package main

import (
	"context"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"

	"github.com/jeffreyyong/news-feeder/internal/app"
	"github.com/jeffreyyong/news-feeder/internal/app/listeners/httplistener"
	"github.com/jeffreyyong/news-feeder/internal/config"
	"github.com/jeffreyyong/news-feeder/internal/logging"
	"github.com/jeffreyyong/news-feeder/internal/transport/transporthttp"
)

const (
	serverServiceName = "news-feeder"
)

var serverCommand = &cli.Command{
	Name:    "server",
	Aliases: []string{"s"},
	Usage:   "Starts the news-feeder http server.",
	Action:  serverAction,
}

func serverAction(ctx *cli.Context) error {
	if err := app.Run(serverServiceName, serverSetup); err != nil {
		logging.From(ctx.Context).Fatal("failed to start service",
			zap.String("service", serverServiceName),
			zap.Error(err),
		)
		return err
	}
	return nil
}

func serverSetup(ctx context.Context, s *app.Service) ([]app.Listener, context.Context, error) {
	s.OnShutdown(func() {
		logging.Print(ctx, "shutdown",
			zap.String("service", serverServiceName),
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

	migrated, version, err := store.Migrate(cfg.MigrationPath)
	if err != nil {
		logging.Error(ctx, "unable to migrate", zap.Uint("version", version))
		return nil, ctx, errors.Wrap(err, "unable to migrate repository")
	}

	if migrated {
		logging.Print(ctx, "postgres database updated", zap.Uint("version", version))
	} else {
		logging.Print(ctx, "postgres database version stayed the same", zap.Uint("version", version))
	}

	feedService, err := newService(ctx, s, cfg, store)
	if err != nil {
		return nil, ctx, errors.Wrap(err, "unable to create feed service")
	}

	socialService, err := newSocialService(ctx, s, cfg, store)
	if err != nil {
		return nil, ctx, errors.Wrap(err, "unable to create social service")
	}

	h, err := transporthttp.NewHTTPHandler(feedService, socialService, transporthttp.WithAuth(cfg.PrivilegedTokens))
	if err != nil {
		logging.Error(ctx, "creating_http_handler", zap.Error(err))
		return nil, ctx, err
	}

	return []app.Listener{httplistener.New(h)}, ctx, nil
}
