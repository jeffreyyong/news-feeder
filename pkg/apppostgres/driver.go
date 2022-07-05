package apppostgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jeffreyyong/news-feeder/internal/app"
)

type ErrInvalidConfig struct {
	Message string
	Field   string
}

func (e ErrInvalidConfig) Error() string {
	return fmt.Sprintf("%s: %s", e.Message, e.Field)
}

// HealthCheckerFunc will query the health of the driver connection.
type healthCheckerFunc func(ctx context.Context) error

// Ping will return the result of the checker func.
func (fn healthCheckerFunc) Ping(ctx context.Context) error { return fn(ctx) }

// NewClient creates a new postgres driver, attaches DataDog monitoring and registers the database with app health checks.
func NewClient(ctx context.Context, app *app.Service, opts ...PGOption) (*sql.DB, error) {
	c := PGConfig{}
	for _, opt := range opts {
		opt(&c)
	}

	postgresDSN := c.dsn

	db, err := NewBasicClient(
		ctx,
		app.Name(),
		postgresDSN,
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}
