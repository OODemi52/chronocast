package routes

import (
	"net/http"

	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
)

func SetupRoutes(rtmpServer *rtmpserver.SimpleRealtimeServer) *http.ServeMux {

	muxRouter := http.NewServeMux()

	SetupHealthRoutes(muxRouter)

	SetupAuthRoutes(muxRouter)

	SetupAPIRoutes(muxRouter, rtmpServer)

	return muxRouter

}
