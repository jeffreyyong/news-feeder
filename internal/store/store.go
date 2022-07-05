package store

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate/v4"
	postgresMigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

const (
	table = "news_feeder"
)

var (
	psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	ErrInvalidParam = errors.New("invalid_parameter")
)

type connKey struct{}

type conn interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Store represents a data repository backed by PostgreSQL.
type Store struct {
	db *sqlx.DB
}

// New creates and returns a new instance of the Store.
func New(db *sqlx.DB) (*Store, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: db", ErrInvalidParam)
	}
	return &Store{db: db}, nil
}

func (s Store) Migrate(path string) (updated bool, version uint, err error) {
	driver, err := postgresMigrate.WithInstance(s.db.DB, &postgresMigrate.Config{
		MigrationsTable: fmt.Sprintf("%s_%s", table, postgresMigrate.DefaultMigrationsTable),
	})
	if err != nil {
		return false, version, errors.Wrap(err, "with instance migration failed")
	}

	// Read migration files
	var m *migrate.Migrate
	m, err = migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", path), "postgres", driver)
	if err != nil {
		return false, version, errors.Wrap(err, "unable to create migrate instance")
	}
	// Perform database migration
	if migrateErr := m.Up(); migrateErr != nil {
		if errors.Is(migrateErr, migrate.ErrNoChange) {
			v, _, _ := m.Version()
			return false, v, nil
		}
		return false, version, errors.Wrap(migrateErr, "unable to up migrations")
	}
	v, _, _ := m.Version()
	return true, v, nil
}

// ExecInTransaction will execute a db function within an atomic sql transaction and rollback on any errors.
// If a connection already exists in the context, this will be reused.
func (s Store) ExecInTransaction(ctx context.Context, f func(ctx context.Context) error) error {
	var fn func() error

	if conn, ok := ctx.Value(connKey{}).(conn); conn != nil && ok {
		return f(ctx)
	}

	fn = func() error {
		txn, err := s.db.BeginTxx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
		if err != nil {
			return err
		}
		if err = f(context.WithValue(ctx, connKey{}, txn)); err != nil {
			if rErr := txn.Rollback(); rErr != nil {
				return rErr
			}
			return err
		}
		if err := txn.Commit(); err != nil {
			return err
		}
		return nil
	}
	return RetryOnPostgresError(fn)
}

func (s Store) connFromContext(ctx context.Context) conn {
	c := ctx.Value(connKey{})
	if conn, ok := c.(conn); conn != nil && ok {
		return conn
	}
	return s.db
}
