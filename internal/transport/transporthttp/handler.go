package transporthttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/jeffreyyong/news-feeder/internal/app/listeners/httplistener"
	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/jeffreyyong/news-feeder/internal/logging"
)

const (
	EndpointListArticles = "/feeds"

	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

type Service interface {
	ListArticles(ctx context.Context, f *domain.SelectArticleFilters) ([]*domain.Article, error)
	ListFeeds(ctx context.Context, f *domain.SelectFeedFilters) ([]*domain.Feed, error)
}

// httpHandler is the http handler that will enable
// calls to this service via HTTP REST
type httpHandler struct {
	service         Service
	middlewareFuncs []mux.MiddlewareFunc
}

// NewHTTPHandler will create a new instance of httpHandler
func NewHTTPHandler(service Service, opts ...MiddlewareFunc) (*httpHandler, error) {
	if service == nil {
		return nil, fmt.Errorf("%w: service", errors.New("some error"))
	}

	h := &httpHandler{service: service}
	for _, opt := range opts {
		if err := opt(h); err != nil {
			return nil, err
		}
	}

	return h, nil
}

// ApplyRoutes will link the HTTP REST endpoint to the corresponding function in this handler
func (h *httpHandler) ApplyRoutes(m *httplistener.Mux) {
	m.HandleFunc(EndpointListArticles, h.ListFeeds).Methods(http.MethodGet)
	m.Use(h.middlewareFuncs...)
}

func (h *httpHandler) ListFeeds(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// get query params

	w.Header().Add(ContentType, ApplicationJSON)
	feed := &Feed{Title: "ffobar"}
	err := json.NewEncoder(w).Encode(feed)
	if err != nil {
		errMsg := "error encoding json response"
		logging.Error(ctx, errMsg, zap.Error(err))
		_ = WriteError(w, errMsg, CodeUnknownFailure)
		return
	}
}
