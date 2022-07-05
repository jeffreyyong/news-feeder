package store

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/lib/pq"
)

var createArticleSQLErrors = map[string]error{
	"article_pkey": domain.ErrArticleAlreadyExists,
}

func (s Store) CreateArticle(ctx context.Context, article *domain.Article) error {
	clauses := map[string]interface{}{
		"title":        article.Title,
		"description":  article.Description,
		"link":         article.Link,
		"tumbnail_url": article.ThumbnailURL,
		"updated_at":   article.UpdatedAt,
	}

	query, args, err := psql.
		Insert("article").
		SetMap(clauses).
		Suffix(`RETURNING id, created_at`).
		ToSql()
	if err != nil {
		return err
	}

	if err = s.connFromContext(ctx).GetContext(ctx, article, query, args...); err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if mappedErr, ok := createArticleSQLErrors[pqErr.Constraint]; ok {
				return mappedErr
			}
		}
		return err
	}
	return nil
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
