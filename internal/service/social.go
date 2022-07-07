package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/jeffreyyong/news-feeder/internal/domain"
)

type TwitterClient interface {
	ShareTweet(ctx context.Context, article string) error
}

type SocialService struct {
	store         Store
	twitterClient TwitterClient
}

func NewSocialService(store Store, twitter TwitterClient) (*SocialService, error) {
	if store == nil {
		return nil, errors.New("nil store")
	}

	return &SocialService{store: store, twitterClient: twitter}, nil
}

func (s *SocialService) Share(ctx context.Context, articleLink string, medium domain.Medium) error {
	switch medium {
	case domain.MediumTwitter:
		if err := s.twitterClient.ShareTweet(ctx, articleLink); err != nil {
			return fmt.Errorf("error sending tweet: %w", err)
		}
	}
	return nil
}
