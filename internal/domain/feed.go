package domain

import (
	"errors"
	"time"

	uuid "github.com/kevinburke/go.uuid"
)

var (
	ErrFeedAlreadyExists = errors.New("feed already exists")
)

type Feed struct {
	ID        uuid.UUID `db:"id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`

	Title       string `db:"title"`
	Description string `db:"description"`
	Link        string `db:"link"`
	FeedLink    string `db:"feed_link"`
	Category    string `db:"category"`
	Language    string `db:"language"`
	Provider    string `db:"provider"`

	Articles []*Article
}

type SelectFeedFilters struct {
	Categories []string
	Providers  []string
	Limit      *uint64
	Offset     *uint64
}
