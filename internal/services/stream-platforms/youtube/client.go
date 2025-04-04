package youtube

import (
	"context"
	"fmt"
	"os"

	"golang.org/x/oauth2"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"

	"github.com/OODemi52/chronocast-server/internal/config"
)

type Client struct {
	oauthConfig *oauth2.Config
}

func NewClient() (*Client, error) {

	oauthConfig, err := config.GetOAuthConfig("youtube")

	if err != nil {
		return nil, fmt.Errorf("failed to get YouTube OAuth config: %w", err)
	}

	return &Client{
		oauthConfig: oauthConfig,
	}, nil

}

func (c *Client) Authenticate(ctx context.Context, code string) (*oauth2.Token, error) {

	token, err := c.oauthConfig.Exchange(ctx, code)

	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}

	//TODO - Store token securely in db for production

	// Store the token in a file named "token.txt" in the root directory and return it
	filePath := "token.txt"
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create token file: %w", err)
	}
	defer file.Close()
	_, err = file.WriteString(token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to write token to file: %w", err)
	}

	return token, nil

}

func (c *Client) GetYouTubeService(ctx context.Context, accessToken string) (*youtube.Service, error) {

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: accessToken,
	})

	service, err := youtube.NewService(ctx, option.WithTokenSource(tokenSource))

	return service, err
}

func (c *Client) GetAccessToken() (string, error) {
	tokenBytes, err := os.ReadFile("token.txt")
	if err != nil {
		return "", fmt.Errorf("failed to read token: %v", err)
	}
	return string(tokenBytes), nil
}
