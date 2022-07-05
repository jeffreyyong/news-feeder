package main

import (
	"context"
	"os"
	"path"

	"go.uber.org/zap"

	"github.com/jeffreyyong/news-feeder/internal/app"
	"github.com/jeffreyyong/news-feeder/internal/app/listeners/httplistener"
	"github.com/jeffreyyong/news-feeder/internal/config"
	"github.com/jeffreyyong/news-feeder/internal/crawler"
	"github.com/jeffreyyong/news-feeder/internal/rss"
	"github.com/jeffreyyong/news-feeder/internal/service"
	"github.com/jeffreyyong/news-feeder/internal/store"
	transporthttp "github.com/jeffreyyong/news-feeder/internal/transport/transporthttp"
	"github.com/jeffreyyong/news-feeder/logging"
	"github.com/jeffreyyong/news-feeder/pkg/apppostgres"
	"github.com/jmoiron/sqlx"

	"github.com/jonboulle/clockwork"
	"github.com/pkg/errors"
)

const (
	serviceName = "news-feeder"

	defaultMigrationPath = "/migrations"
)

func main() {
	if err := app.Run(serviceName, setup); err != nil {
		logging.Error(context.Background(), "failed to start service",
			zap.String("service", serviceName),
			zap.Error(err))
		panic(err)
	}
}

func setup(ctx context.Context, s *app.Service) ([]app.Listener, context.Context, error) {
	s.OnShutdown(func() {
		logging.Print(ctx, "shutdown",
			zap.String("service", serviceName),
		)
	})

	cfg, err := config.Load()
	if err != nil {
		logging.Error(ctx, "loading_config", zap.Error(err))
		return nil, ctx, err
	}

	store, err := store.New(cfg.PostgresDSN)
	if err != nil {
		logging.Error(ctx, "initialising store", zap.Error(err))
		return nil, ctx, errors.Wrap(err, "initialising store")
	}
	migrationPath, err := migrationPath()
	if err != nil {
		logging.Error(ctx, "unable to get migration path", zap.Error(err))
		return nil, ctx, errors.Wrap(err, "unable to get migration path")
	}
	if err = store.Migrate(migrationPath); err != nil {
		logging.Error(ctx, "unable to migrate")
		return nil, ctx, errors.Wrap(err, "unable to migrate repository")
	}

	svc, err := service.NewService(store, service.WithClock(clockwork.NewRealClock()))

	if err != nil {
		logging.Error(ctx, "creating_service", zap.Error(err))
		return nil, ctx, err
	}

	h, err := transporthttp.NewHTTPHandler(svc, transporthttp.WithAuth(cfg.PrivilegedTokens))
	if err != nil {
		logging.Error(ctx, "creating_http_handler", zap.Error(err))
		return nil, ctx, err
	}

	return []app.Listener{httplistener.New(h)}, ctx, nil
}

func migrationPath() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return path.Join(wd, defaultMigrationPath), nil
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
