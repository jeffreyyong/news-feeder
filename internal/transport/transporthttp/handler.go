package transporthttp

import "context"

type Service interface {
	ListArticles(ctx context.Context) (*domain.Article, error)
}
