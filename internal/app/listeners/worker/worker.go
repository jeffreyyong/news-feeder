package worker

import (
	"context"
	"time"

	"github.com/jeffreyyong/news-feeder/internal/logging"
)

type Service interface {
	CrawlFeeds(ctx context.Context) error
}

type Worker struct {
	interval time.Duration
	service  Service

	ctxCancel func()
}

func New(service Service, interval time.Duration) *Worker {
	return &Worker{
		service:  service,
		interval: interval,
	}
}

func (w *Worker) Serve(ctx context.Context) error {
	logging.Print(ctx, "starting worker")
	logging.Print(ctx, w.interval.String())

	ctx, w.ctxCancel = context.WithCancel(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logging.Print(ctx, "stopped worker")
			return nil
		case <-ticker.C:
			logging.Print(ctx, "ticking")
			// if err := w.service.CrawlFeeds(ctx); err != nil {
			// 	logging.Error(ctx, "failed to fetch articles", zap.Error(err))
			// }
		}
	}
}

func (w *Worker) Close(ctx context.Context) error {
	logging.Print(ctx, "stop worker")
	w.ctxCancel()
	return nil
}

func (w *Worker) Name() string {
	return "crawler_woker"
}
