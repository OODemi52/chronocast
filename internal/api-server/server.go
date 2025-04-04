package apiserver

import (
	"context"
	"log"
	"net/http"

	"github.com/OODemi52/chronocast-server/internal/api-server/routes"
	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
)

type APIServer struct {
	httpServer *http.Server
	port       string
}

func NewServer(port string, rtmpServer *rtmpserver.NginxServer) (*APIServer, error) {

	return &APIServer{
		port: port,
		httpServer: &http.Server{
			Addr:    port,
			Handler: routes.SetupRoutes(rtmpServer),
		},
	}, nil

}

func (as *APIServer) Start() error {

	log.Printf("HTTP server starting on %s...", as.port)

	err := as.httpServer.ListenAndServe()

	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil

}

func (as *APIServer) Shutdown(ctx context.Context) error {

	log.Printf("Shutting down HTTP server...")

	return as.httpServer.Shutdown(ctx)

}
