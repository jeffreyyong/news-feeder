package rss

import (
	"context"
	"errors"
	"strings"
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

	var articles []*domain.Article

	for _, i := range f.Items {

		var thumbnailURL string
		for _, enclosure := range i.Enclosures {
			if enclosure.Type == EnclosureTypeImage {
				thumbnailURL = enclosure.URL
			}
		}

		var publishedAt time.Time
		if i.PublishedParsed != nil {
			publishedAt = *i.PublishedParsed
		}

		article := &domain.Article{
			PublishedAt:  publishedAt,
			Title:        i.Title,
			Description:  i.Description,
			Link:         i.Link,
			ThumbnailURL: thumbnailURL,
			GUID:         i.GUID,
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
		FeedLink:    url,
		Category:    mapCategory(f.FeedLink),
		Language:    strings.ToLower(f.Language),
		UpdatedAt:   updatedAt,
		Articles:    articles,
		Provider:    mapProvider(f.FeedLink),
	}

	return feed, nil
}

func mapCategory(title string) domain.Category {
	t := strings.ToLower(title)
	switch {
	case strings.Contains(t, "uk"):
		return domain.CategoryUK
	case strings.Contains(t, "technology"):
		return domain.CategoryTechnology
	default:
		return domain.CategoryUnknown
	}
}

func mapProvider(feedLink string) domain.Provider {
	l := strings.ToLower(feedLink)
	switch {
	case strings.Contains(l, "sky"):
		return domain.ProviderSky
	case strings.Contains(l, "bbc"):
		return domain.ProviderBBC
	default:
		return domain.ProviderUnknown
	}
}
