package domain

import (
	"errors"
	"time"

	uuid "github.com/kevinburke/go.uuid"
)

var (
	ErrArticleAlreadyExists = errors.New("article already exists")
)

type Article struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	FeedID      uuid.UUID  `db:"feed_id" json:"-"`
	PublishedAt time.Time  `db:"published_at" json:"published_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   *time.Time `db:"updated_at" json:"updated_at"`

	Title        string `db:"title" json:"title"`
	Description  string `db:"description" json:"description"`
	Link         string `db:"link" json:"link"`
	ThumbnailURL string `db:"thumbnail_url" json:"thumbnail_url"`
	GUID         string `db:"guid" json:"-"`
}

type SelectArticleFilters struct {
	Limit      *uint64
	Offset     *uint64
	Categories []Category
	Providers  []Provider
}
