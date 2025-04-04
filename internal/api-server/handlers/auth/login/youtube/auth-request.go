package handlers

import (
	"net/http"

	"github.com/OODemi52/chronocast-server/internal/config"
	"github.com/OODemi52/chronocast-server/internal/utils"
	"golang.org/x/oauth2"
)

func OAuthLoginRequestHandler(w http.ResponseWriter, r *http.Request) {
	platform := utils.GetPlatformFromRequest(w, r)

	oauth2Config, err := config.GetOAuthConfig(platform)

	if err != nil {
		http.Error(w, "Failed to get OAuth configuration", http.StatusInternalServerError)
		return
	}

	state := utils.GetStateString(w)

	platformCookie := &http.Cookie{
		Name:     "platform",
		Value:    platform,
		HttpOnly: true,
		Path:     "/",
	}
	http.SetCookie(w, platformCookie)

	stateCookie := &http.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HttpOnly: true,
		Secure:   false, //TODO - Set to true later, or dynamically with env, when using https
		Path:     "/",
	}
	http.SetCookie(w, stateCookie)

	authURL := oauth2Config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	http.Redirect(w, r, authURL, http.StatusFound)
}
