package routes

import (
	"net/http"

	youtubeLoginHandlers "github.com/OODemi52/chronocast-server/internal/api-server/handlers/auth/login/youtube"
	"github.com/OODemi52/chronocast-server/internal/api-server/middleware"
)

func SetupAuthRoutes(mux *http.ServeMux) {

	mux.Handle("/auth/login/youtube", middleware.ChainMiddleware(
		http.HandlerFunc(youtubeLoginHandlers.OAuthLoginRequestHandler),
		middleware.Logging,
		middleware.CORS,
	))

	mux.Handle("/auth/login/youtube/callback", middleware.ChainMiddleware(
		http.HandlerFunc(youtubeLoginHandlers.OAuthLoginCallbackHandler),
		middleware.Logging,
		middleware.CORS,
	))

	mux.Handle("/auth/login/youtube/revoke", middleware.ChainMiddleware(
		http.HandlerFunc(youtubeLoginHandlers.RevokeOAuthLoginTokenHandler),
		middleware.Logging,
		middleware.CORS,
	))

	mux.Handle("/auth/login/youtube/refresh", middleware.ChainMiddleware(
		http.HandlerFunc(youtubeLoginHandlers.RefreshOAuthLoginTokenHandler),
		middleware.Logging,
		middleware.CORS,
	))

}
