package store

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/lib/pq"
)

var createArticleSQLErrors = map[string]error{
	"article_pkey": domain.ErrArticleAlreadyExists,
}

func (s Store) CreateArticle(ctx context.Context, article *domain.Article) (string, error) {
	clauses := map[string]interface{}{
		"title":         article.Title,
		"description":   article.Description,
		"link":          article.Link,
		"thumbnail_url": article.ThumbnailURL,
		"updated_at":    article.UpdatedAt,
		"published_at":  article.PublishedAt,
		"feed_id":       article.FeedID,
		"guid":          article.GUID,
	}

	query, args, err := psql.
		Insert("article").
		SetMap(clauses).
		Suffix(`ON CONFLICT (guid) DO UPDATE SET guid = excluded.guid RETURNING id`).
		ToSql()
	if err != nil {
		return "", err
	}

	row := s.connFromContext(ctx).QueryRowxContext(ctx, query, args...)
	if row.Err() != nil {
		if pqErr, ok := row.Err().(*pq.Error); ok {
			if mappedErr, ok := createArticleSQLErrors[pqErr.Constraint]; ok {
				return "", mappedErr
			}
		}
		return "", row.Err()
	}

	var id string
	err = row.Scan(&id)
	if err != nil {
		return "", fmt.Errorf("failed to return article id: %w", err)
	}
	return id, nil
}

func applySelectArticleFilters(f *domain.SelectArticleFilters, query sq.SelectBuilder) sq.SelectBuilder {
	if len(f.Categories) > 0 {
		query = query.Where(sq.Eq{"feed.category": f.Categories})
	}

	if len(f.Providers) > 0 {
		query = query.Where(sq.Eq{"feed.provider": f.Providers})
	}

	if f.Limit != nil {
		query = query.Limit(*f.Limit)
	}

	if f.Offset != nil {
		query = query.Offset(*f.Offset)
	}

	return query
}

func (s Store) SelectArticles(ctx context.Context, f *domain.SelectArticleFilters) ([]*domain.Article, error) {
	queryBuilder := psql.Select().
		Columns(
			"article.id as id",
			"article.title as title",
			"article.description as description",
			"article.thumbnail_url as thumbnail_url",
			"article.created_at as created_at",
			"article.updated_at as updated_at",
			"article.published_at as published_at",
		).
		From("article").
		LeftJoin("feed ON article.feed_id = feed.id").
		OrderBy("published_at DESC")

	if f != nil {
		queryBuilder = applySelectArticleFilters(f, queryBuilder)
	}
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	var articles []*domain.Article
	if err = s.connFromContext(ctx).SelectContext(ctx, &articles, query, args...); err != nil {
		return nil, err
	}
	return articles, nil
}
