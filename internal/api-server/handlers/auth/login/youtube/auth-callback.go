package handlers

import (
	"fmt"
	"net/http"
	"os"

	"github.com/OODemi52/chronocast-server/internal/config"
	"github.com/OODemi52/chronocast-server/internal/utils"
)

func OAuthLoginCallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := utils.GetCodeFromRequest(w, r)

	stateCookie, err := r.Cookie("oauthstate")

	if err != nil {
		http.Error(w, "State cookie not found", http.StatusBadRequest)
		return
	}

	storedState := stateCookie.Value

	queryState := r.URL.Query().Get("state")

	if queryState == "" {
		http.Error(w, "State parameter missing", http.StatusBadRequest)
		return
	}

	if storedState != queryState {
		http.Error(w, "Invalid OAuth state", http.StatusBadRequest)
		return
	}

	platformCookie, err := r.Cookie("platform")

	if err != nil {
		http.Error(w, "Platform cookie not found", http.StatusBadRequest)
		return
	}

	platform := platformCookie.Value

	oauth2Config, err := config.GetOAuthConfig(platform)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get OAuth config: %v", err), http.StatusInternalServerError)
		return
	}

	token, err := oauth2Config.Exchange(r.Context(), code)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code for token: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "YouTube authenticated successfully! Access Token: %s", token.AccessToken)

	//TODO - Securely store accesstoken, as well as expire and refresh logic
	// For development: save the token to a text file.
	err = os.WriteFile("token.txt", []byte(token.AccessToken), 0644)

	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to write token to file: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "http://localhost:3000", http.StatusSeeOther)
}
