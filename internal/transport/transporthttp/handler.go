package transporthttp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/jeffreyyong/news-feeder/internal/app/listeners/httplistener"
	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/jeffreyyong/news-feeder/internal/logging"
)

const (
	EndpointListArticles = "/articles"
	EndpointShareArticle = "/article/share"

	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

type FeedService interface {
	ListArticles(ctx context.Context, f *domain.SelectArticleFilters) ([]*domain.Article, error)
	ListFeeds(ctx context.Context, f *domain.SelectFeedFilters) ([]*domain.Feed, error)
}

type SocialService interface {
	Share(ctx context.Context, articleLink string, medium domain.Medium) error
}

// httpHandler is the http handler that will enable
// calls to this service via HTTP REST
type httpHandler struct {
	feedService     FeedService
	socialService   SocialService
	middlewareFuncs []mux.MiddlewareFunc
}

// NewHTTPHandler will create a new instance of httpHandler
func NewHTTPHandler(feedService FeedService, socialService SocialService, opts ...MiddlewareFunc) (*httpHandler, error) {
	if feedService == nil {
		return nil, fmt.Errorf("nil feed service")
	}

	if socialService == nil {
		return nil, fmt.Errorf("nil social service")
	}

	h := &httpHandler{feedService: feedService, socialService: socialService}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

// ApplyRoutes will link the HTTP REST endpoint to the corresponding function in this handler
func (h *httpHandler) ApplyRoutes(m *httplistener.Mux) {
	m.HandleFunc(EndpointListArticles, h.ListArticles).Methods(http.MethodGet)
	m.HandleFunc(EndpointShareArticle, h.ShareArticle).Methods(http.MethodPost)
	m.Use(h.middlewareFuncs...)
}

// ShareArticle allows the client to share an interesting article link
// Currently supported medium is Twitter.
func (h *httpHandler) ShareArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	type ReqBody struct {
		Medium      string `json:"medium"`
		ArticleLink string `json:"link"`
	}

	var reqBody ReqBody

	err := json.NewDecoder(r.Body).Decode(&reqBody)
	if err != nil {
		errMsg := "bad request body"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeBadRequest)
		return
	}

	medium := domain.Medium(reqBody.Medium)

	if _, ok := domain.SupportedMedium[medium]; !ok {
		errMsg := "unsupported sharing medium"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeBadRequest)
		return
	}

	err = h.socialService.Share(ctx, reqBody.ArticleLink, medium)
	if err != nil {
		errMsg := "error sharing article"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeUnknownFailure)
		return
	}
}

// ListArticles allows the client to list the articles by "categories" and "providers".
// Pagination is also supported by providing "limit" and "offset"
// Example: GET /articles?categories=uk,technology&providers=bbc
func (h *httpHandler) ListArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get query params
	categoryQuery := r.URL.Query().Get("categories")
	var domainCategories []domain.Category
	var err error
	if categoryQuery != "" {
		categories := strings.Split(categoryQuery, ",")
		domainCategories, err = mapCategory(categories)
		if err != nil {
			errMsg := "bad query params"
			logging.Error(ctx, errMsg, zap.Error(err))
			_ = WriteError(w, errMsg, CodeBadRequest)
			return
		}
	}

	providerQuery := r.URL.Query().Get("providers")
	var domainProviders []domain.Provider
	if providerQuery != "" {
		providers := strings.Split(providerQuery, ",")
		domainProviders, err = mapProvider(providers)
		if err != nil {
			errMsg := "bad query params"
			logging.Error(ctx, errMsg, zap.Error(err))
			_ = WriteError(w, errMsg, CodeBadRequest)
			return
		}
	}

	limit := r.URL.Query().Get("limit")
	var limitInt uint64
	if limit != "" {
		limitInt, err = strconv.ParseUint(limit, 10, 64)
		if err != nil {
			errMsg := "bad query params"
			logging.Error(ctx, errMsg, zap.Error(err))
			_ = WriteError(w, errMsg, CodeBadRequest)
			return
		}
	}

	offset := r.URL.Query().Get("offset")
	var offsetInt uint64
	if offset != "" {
		offsetInt, err = strconv.ParseUint(offset, 10, 64)
		if err != nil {
			errMsg := "bad query params"
			logging.Error(ctx, errMsg, zap.Error(err))
			_ = WriteError(w, errMsg, CodeBadRequest)
			return
		}
	}

	selectArticlesFilter := &domain.SelectArticleFilters{
		Categories: domainCategories,
		Providers:  domainProviders,
	}

	if limitInt != 0 {
		selectArticlesFilter.Limit = &limitInt
	}

	if offsetInt != 0 {
		selectArticlesFilter.Offset = &offsetInt
	}

	articles, err := h.feedService.ListArticles(ctx, selectArticlesFilter)
	if err != nil {
		errMsg := "error getting articles"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeUnknownFailure)
		return
	}

	w.Header().Add(ContentType, ApplicationJSON)
	err = json.NewEncoder(w).Encode(articles)
	if err != nil {
		errMsg := "error encoding json response"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeUnknownFailure)
		return
	}
}

func mapCategory(categories []string) ([]domain.Category, error) {
	domainCategories := make([]domain.Category, len(categories))

	for _, c := range categories {
		category := domain.Category(c)
		if _, ok := domain.SupportedCategory[category]; !ok {
			return nil, fmt.Errorf("unsupported category: %s", category)
		}
		domainCategories = append(domainCategories, category)
	}
	return domainCategories, nil
}

func mapProvider(providers []string) ([]domain.Provider, error) {
	domainProviders := make([]domain.Provider, len(providers))

	for _, c := range providers {
		provider := domain.Provider(c)
		if _, ok := domain.SupportedProvider[provider]; !ok {
			return nil, fmt.Errorf("unsupported provider: %s", provider)
		}
		domainProviders = append(domainProviders, provider)
	}
	return domainProviders, nil
}
