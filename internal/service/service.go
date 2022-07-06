package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/jeffreyyong/news-feeder/internal/logging"
	"github.com/jonboulle/clockwork"

	uuid "github.com/kevinburke/go.uuid"
)

// Store is the db interface
type Store interface {
	ExecInTransaction(ctx context.Context, f func(ctx context.Context) error) error

	CreateFeed(ctx context.Context, feed *domain.Feed) (string, error)
	SelectFeeds(ctx context.Context, f *domain.SelectFeedFilters) ([]*domain.Feed, error)

	CreateArticle(ctx context.Context, article *domain.Article) (string, error)
	SelectArticles(ctx context.Context, f *domain.SelectArticleFilters) ([]*domain.Article, error)
}

type Crawler interface {
	Crawl(ctx context.Context) ([]*domain.Feed, error)
}

type Service struct {
	store   Store
	clock   clockwork.Clock
	crawler Crawler
}

func New(store Store, crawler Crawler, opts ...Option) (*Service, error) {
	if store == nil {
		return nil, errors.New("nil store")
	}

	s := &Service{store: store, crawler: crawler}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	return s, nil
}

// ListArticles lists articles that have been stored in the persistence layer.
func (s *Service) ListArticles(ctx context.Context, filters *domain.SelectArticleFilters) ([]*domain.Article, error) {
	return nil, nil
}

// ListFeeds lists feeds that have been stored in the persistence layer.
func (s *Service) ListFeeds(ctx context.Context, filters *domain.SelectFeedFilters) ([]*domain.Feed, error) {
	return nil, nil
}

// 1. take the sources, for each feed, create Feed and create Article.
func (s *Service) CrawlFeeds(ctx context.Context) error {
	feeds, err := s.crawler.Crawl(ctx)
	if err != nil {
		return err
	}

	logging.Print(ctx, "finished crawling")

	if err := s.store.ExecInTransaction(ctx, func(ctx context.Context) error {
		for _, feed := range feeds {
			feedID, err := s.store.CreateFeed(ctx, feed)
			if err != nil {
				return fmt.Errorf("error creating feed in db: %w", err)
			}

			for _, article := range feed.Articles {
				id, _ := uuid.FromString(feedID)
				article.FeedID = id
				logging.Print(ctx, "what is the feed ID")
				logging.Print(ctx, article.FeedID.String())
				_, err := s.store.CreateArticle(ctx, article)
				if err != nil {
					return fmt.Errorf("error creating article in db: %w", err)
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}
