package crawler

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/jeffreyyong/news-feeder/internal/domain"
)

type Parser interface {
	Parse(ctx context.Context, url string) (*domain.Feed, error)
}

const (
	SleepDuration = 100 * time.Millisecond
)

func New(parser Parser, sources []string) *Crawler {
	return &Crawler{
		feedParser: parser,
		sources:    sources,
	}
}

type Crawler struct {
	sources    []string
	feedParser Parser
}

func (c *Crawler) Crawl(ctx context.Context) ([]*domain.Feed, error) {
	results := make(chan *domain.Feed)
	g, ctx := errgroup.WithContext(ctx)

	for _, s := range c.sources {
		source := s
		g.Go(func() error {
			feed, err := c.feedParser.Parse(ctx, source)
			if err != nil {
				return fmt.Errorf("error parsing url (%s): %w", source, err)
			}
			results <- feed
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	var feeds []*domain.Feed

	for result := range results {
		feeds = append(feeds, result)
	}
	return feeds, nil
}
