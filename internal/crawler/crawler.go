package crawler

import (
	"context"
	"fmt"

	"github.com/jeffreyyong/news-feeder/internal/domain"
)

type Parser interface {
	Parse(ctx context.Context, url string) (*domain.Feed, error)
}

func New(parser Parser, sources []string) *Crawler {
	return &Crawler{
		feedParser: parser,
		sources:    sources,
		results:    make(chan *domain.Feed),
	}
}

type Crawler struct {
	sources    []string
	feedParser Parser
	results    chan *domain.Feed
}

func (c *Crawler) Crawl(ctx context.Context) ([]*domain.Feed, error) {
	var feeds []*domain.Feed

	for _, s := range c.sources {
		feed, err := c.feedParser.Parse(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("error parsing url (%s): %w", s, err)
		}

		feeds = append(feeds, feed)
	}

	return feeds, nil
}
