package routes

import (
	"net/http"

	apiHandlers "github.com/OODemi52/chronocast-server/internal/api-server/handlers/api"
	"github.com/OODemi52/chronocast-server/internal/api-server/middleware"
	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
)

func SetupAPIRoutes(mux *http.ServeMux, rtmpServer *rtmpserver.SimpleRealtimeServer) {

	mux.HandleFunc("/api/rtmp/auth", apiHandlers.RTMPAuthHandler)

	mux.HandleFunc("/api/rtmp/publish-done", apiHandlers.RTMPPublishDoneHandler)

	mux.Handle("/api/generate-stream-key", middleware.ChainMiddleware(
		http.HandlerFunc(apiHandlers.GenerateStreamKeyHandler),
		middleware.CORS,
		middleware.Logging,
	))

	mux.Handle("/api/streams", middleware.ChainMiddleware(
		apiHandlers.CreateStreamHandler(rtmpServer),
		middleware.CORS,
		middleware.Logging,
	))

	mux.Handle("/api/streams/", middleware.ChainMiddleware(
		apiHandlers.ManageStreamHandler(rtmpServer),
		middleware.CORS,
		middleware.Logging,
	))
}
