package domain

import (
	"database/sql"
	"errors"
	"time"

	uuid "github.com/kevinburke/go.uuid"
)

var (
	ErrArticleAlreadyExists = errors.New("article already exists")
)

type Article struct {
	ID          uuid.UUID    `db:"id"`
	FeedID      uuid.UUID    `db:"feed_ud"`
	PublishedAt sql.NullTime `db:"published_at"`
	CreatedAt   time.Time    `db:"created_at"`
	UpdatedAt   sql.NullTime `db:"updated_at"`

	Title        string `db:"title"`
	Description  string `db:"description"`
	Link         string `db:"link"`
	ThumbnailURL string `db:"thumbnail_url"`
}

type Category string

const (
	CategorySports Category = "SPORTS"
)

type Provider string

const (
	ProviderBBC Provider = "BBC"
)

type SelectArticleFilters struct {
	Limit  *uint64
	Offset *uint64
}
