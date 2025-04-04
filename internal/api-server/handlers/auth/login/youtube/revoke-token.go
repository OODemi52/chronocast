package handlers

import (
	"fmt"
	"net/http"

	"golang.org/x/oauth2"
)

func RevokeOAuthLoginTokenHandler(w http.ResponseWriter, r *http.Request) {
	accessToken := r.URL.Query().Get("access_token")

	if accessToken == "" {
		http.Error(w, "Missing access token", http.StatusBadRequest)
		return
	}

	token := &oauth2.Token{AccessToken: accessToken}

	revokeURL := fmt.Sprintf("https://oauth2.googleapis.com/revoke?token=%s", token.AccessToken)

	resp, err := http.Get(revokeURL)

	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to revoke token", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Token revoked successfully!")
}
