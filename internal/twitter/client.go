package twitter

import (
	"context"
	"fmt"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
)

type Creds struct {
	ConsumerKey    string
	ConsumerSecret string
	AccessToken    string
	AccessSecret   string
}

type Client struct {
	*twitter.Client
}

func NewClient(c Creds) (*Client, error) {
	if c.ConsumerKey == "" || c.ConsumerSecret == "" || c.AccessToken == "" || c.AccessSecret == "" {
		return nil, fmt.Errorf("Consumer key/secret and Access token/secret required")
	}
	config := oauth1.NewConfig(c.ConsumerKey, c.ConsumerSecret)
	token := oauth1.NewToken(c.AccessToken, c.AccessSecret)

	httpClient := config.Client(context.Background(), token)

	client := twitter.NewClient(httpClient)

	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	_, _, err := client.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		return nil, fmt.Errorf("failed to verify credentials: %w", err)
	}

	return &Client{Client: client}, nil
}

func (c *Client) ShareTweet(ctx context.Context, articleLink string) error {
	tweetMsg := fmt.Sprintf("This is an interesting read: %s", articleLink)
	_, _, err := c.Statuses.Update(tweetMsg, nil)
	if err != nil {
		return fmt.Errorf("failed to share a tweet: %w", err)
	}
	return nil
}
