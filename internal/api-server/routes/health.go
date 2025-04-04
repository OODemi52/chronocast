package routes

import (
	"net/http"

	healthHandlers "github.com/OODemi52/chronocast-server/internal/api-server/handlers/health"
	"github.com/OODemi52/chronocast-server/internal/api-server/middleware"
)

func SetupHealthRoutes(mux *http.ServeMux) {

	mux.Handle("/health", middleware.ChainMiddleware(
		http.HandlerFunc(healthHandlers.HealthCheckHandler),
		middleware.Logging,
		middleware.CORS,
		middleware.HandleAuth,
	))

	//TODO - Improve health checking
}
