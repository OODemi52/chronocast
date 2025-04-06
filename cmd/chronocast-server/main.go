package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiserver "github.com/OODemi52/chronocast-server/internal/api-server"
	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
	"github.com/joho/godotenv"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found, using environment variables")
	}

	apiPort := flag.String("http-port", ":8080", "HTTP Server port")

	rtmpPort := flag.String("rtmp-port", ":1935", "RTMP Server port")

	flag.Parse()

	rtmpServer, err := rtmpserver.NewServer(*rtmpPort)

	if err != nil {
		log.Fatalf("Failed to initialize RTMP server: %v", err)
	}

	go func() {

		if err := rtmpServer.Start(); err != nil {
			log.Fatalf("RTMP server failed to start: %v", err)
		}

	}()

	apiServer, err := apiserver.NewServer(*apiPort, rtmpServer)

	if err != nil {
		log.Fatalf("Failed to initialize API server: %v", err)
	}

	go func() {

		if err := apiServer.Start(); err != nil {
			log.Fatalf("API server failed to start: %v", err)
		}

	}()

	handleServerShutdown(apiServer, rtmpServer)

}

func handleServerShutdown(apiServer *apiserver.APIServer, rtmpServer *rtmpserver.SimpleRealtimeServer) {

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	sig := <-sigChan

	log.Printf("Received signal: %s, initiating shutdown...", sig)

	const shutdownTimeout = 10 * time.Second

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)

	defer shutdownCancel()

	if apiServer != nil {

		log.Println("Shutting down API server...")

		if err := apiServer.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}

	}

	log.Println("Stopping RTMP server...")

	if err := rtmpServer.Stop(); err != nil {
		log.Printf("RTMP server shutdown error: %v", err)
	}

	log.Println("Shutdown complete")

}
