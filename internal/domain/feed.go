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

	Title       string   `db:"title"`
	Description string   `db:"description"`
	Link        string   `db:"link"`
	FeedLink    string   `db:"feed_link"`
	Category    Category `db:"category"`
	Language    string   `db:"language"`
	Provider    Provider `db:"provider"`

	Articles []*Article
}

type SelectFeedFilters struct {
	Categories []string
	Providers  []string
	Limit      *uint64
	Offset     *uint64
}

type Category string

const (
	CategoryTechnology Category = "Technology"
	CategoryUK         Category = "UK"
	CategoryUnknown    Category = "Unknown"
)

var SupportedCategory = map[Category]bool{
	CategoryTechnology: true,
	CategoryUK:         true,
}

type Provider string

const (
	ProviderBBC     Provider = "BBC"
	ProviderSky     Provider = "Sky"
	ProviderUnknown Provider = "Unknown"
)

var SupportedProvider = map[Provider]bool{
	ProviderBBC: true,
	ProviderSky: true,
}
