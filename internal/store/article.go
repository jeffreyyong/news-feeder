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
		"feed_id":       article.FeedID,
	}

	query, args, err := psql.
		Insert("article").
		SetMap(clauses).
		Suffix(`RETURNING id`).
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
			"id",
			"title",
			"description",
			"thumbnail_url",
			"created_at",
			"updated_at",
			"published_at",
		).
		From("article").
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
