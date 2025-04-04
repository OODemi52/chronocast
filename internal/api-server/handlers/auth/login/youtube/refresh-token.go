package handlers

import (
	"fmt"
	"net/http"

	"github.com/OODemi52/chronocast-server/internal/config"
	"github.com/OODemi52/chronocast-server/internal/utils"
	"golang.org/x/oauth2"
)

func RefreshOAuthLoginTokenHandler(w http.ResponseWriter, r *http.Request) {
	platform := utils.GetPlatformFromRequest(w, r)

	refreshToken := r.URL.Query().Get("refresh_token")

	if refreshToken == "" {
		http.Error(w, "Missing refresh token", http.StatusBadRequest)
		return
	}

	oauth2Config, err := config.GetOAuthConfig(platform)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get OAuth config: %v", err), http.StatusInternalServerError)
		return
	}

	tokenSource := oauth2Config.TokenSource(r.Context(), &oauth2.Token{RefreshToken: refreshToken})

	token, err := tokenSource.Token()

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to refresh token: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "New Access Token: %s", token.AccessToken)

	//TODO - Securely store accesstoken, as well as expire and refresh logic
	// For now, just print it to the response for debugging purposes
}
