package store

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/lib/pq"
)

var createFeedSQLErrors = map[string]error{
	"feed_pkey": domain.ErrFeedAlreadyExists,
}

func (s Store) CreateFeed(ctx context.Context, feed *domain.Feed) (string, error) {
	clauses := map[string]interface{}{
		"title":       feed.Title,
		"description": feed.Description,
		"link":        feed.Link,
		"feed_link":   feed.FeedLink,
		"category":    feed.Category,
		"language":    feed.Language,
		"provider":    feed.Provider,
		"updated_at":  feed.UpdatedAt,
	}

	query, args, err := psql.
		Insert("feed").
		SetMap(clauses).
		Suffix(`ON CONFLICT (feed_link) DO UPDATE SET feed_link = excluded.feed_link RETURNING id`).
		ToSql()
	if err != nil {
		return "", err
	}

	row := s.connFromContext(ctx).QueryRowxContext(ctx, query, args...)
	if row.Err() != nil {
		if pqErr, ok := row.Err().(*pq.Error); ok {
			if mappedErr, ok := createFeedSQLErrors[pqErr.Constraint]; ok {
				return "", mappedErr
			}
		}
		return "", row.Err()
	}

	var id string
	err = row.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to return feed id: %w", err)
	}
	return id, nil
}

func applySelectFeedFilters(f *domain.SelectFeedFilters, query sq.SelectBuilder) sq.SelectBuilder {
	if len(f.Categories) > 0 {
		query = query.Where(sq.Eq{"category": f.Categories})
	}

	if len(f.Providers) > 0 {
		query = query.Where(sq.Eq{"provider": f.Providers})
	}

	if f.Limit != nil {
		query = query.Limit(*f.Limit)
	}
	if f.Offset != nil {
		query = query.Offset(*f.Offset)
	}

	return query
}

func (s Store) SelectFeeds(ctx context.Context, f *domain.SelectFeedFilters) ([]*domain.Feed, error) {
	queryBuilder := psql.Select().
		Columns(
			"id",
			"title",
			"description",
			"link",
			"feed_link",
			"category",
			"language",
			"provider",
			"created_at",
			"updated_at",
		).
		From("feed").
		OrderBy("created_at DESC")

	if f != nil {
		queryBuilder = applySelectFeedFilters(f, queryBuilder)
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var feeds []*domain.Feed
	if err = s.connFromContext(ctx).SelectContext(ctx, &feeds, query, args...); err != nil {
		return nil, err
	}
	return feeds, nil
}
