package config

import (
	"fmt"
	"os"

	"golang.org/x/oauth2"
)

func GetOAuthConfig(platform string) (*oauth2.Config, error) {

	switch platform {

	case "youtube":
		return &oauth2.Config{
			ClientID:     os.Getenv("YOUTUBE_WEB_CLIENT_ID"),
			ClientSecret: os.Getenv("YOUTUBE_WEB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("YOUTUBE_REDIRECT_URI"),
			Scopes:       []string{"https://www.googleapis.com/auth/youtube.force-ssl"},
			Endpoint: oauth2.Endpoint{
				AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
				TokenURL: "https://oauth2.googleapis.com/token",
			},
		}, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)

	}

}
