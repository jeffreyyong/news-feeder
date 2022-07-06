package rss

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jeffreyyong/news-feeder/internal/domain"
	"github.com/mmcdole/gofeed"
)

const (
	EnclosureTypeImage = "image/jpeg"
)

type Parser struct {
	*gofeed.Parser
}

func NewParser() *Parser {
	return &Parser{
		Parser: gofeed.NewParser(),
	}
}

func (p *Parser) Parse(ctx context.Context, url string) (*domain.Feed, error) {
	f, err := p.ParseURL(url)
	if err != nil {
		return nil, errors.New("error parsing feed")
	}

	var category string
	if len(f.Categories) > 0 {
		category = f.Categories[0]
	}

	var articles []*domain.Article

	for _, i := range f.Items {

		var thumbnailURL string
		for _, enclosure := range i.Enclosures {
			if enclosure.Type == EnclosureTypeImage {
				thumbnailURL = enclosure.URL
			}
		}

		var publishedAt sql.NullTime
		if i.PublishedParsed != nil {
			publishedAt = sql.NullTime{Time: *i.PublishedParsed}
		}

		article := &domain.Article{
			PublishedAt:  publishedAt,
			Title:        i.Title,
			Description:  i.Description,
			Link:         i.Link,
			ThumbnailURL: thumbnailURL,
		}
		articles = append(articles, article)
	}

	var updatedAt time.Time
	if f.UpdatedParsed != nil {
		updatedAt = *f.UpdatedParsed
	}

	feed := &domain.Feed{
		Title:       f.Title,
		Description: f.Description,
		Link:        f.Link,
		FeedLink:    f.FeedLink,
		Category:    category,
		Language:    f.Language,
		UpdatedAt:   updatedAt,
		Articles:    articles,
	}

	return feed, nil
}
