package rtmpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
)

type StreamDestination struct {
	URL       string
	StreamKey string
}

type Stream struct {
	Key          string
	Destinations []StreamDestination
}

type SimpleRealtimeServer struct {
	Port        string
	ConfigPath  string
	SRSPath     string
	SRSProcess  *os.Process
	Streams     map[string]*Stream
	StreamsLock sync.RWMutex
	ConfigDir   string
}

func NewServer(port string) (*SimpleRealtimeServer, error) {

	configPath := os.Getenv("CONFIG_PATH")

	srsPath := os.Getenv("SRS_PATH")

	if configPath == "" || srsPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH and SRS_PATH must be set")
	}

	return &SimpleRealtimeServer{
		Port:       port,
		ConfigPath: configPath,
		SRSPath:    srsPath,
		Streams:    make(map[string]*Stream),
	}, nil

}

func (srs *SimpleRealtimeServer) Start() error {

	log.Println("RTMP Server started via Docker, checking it's status..")

	apiURL := "http://localhost:1985/api/v1/versions"

	resp, err := http.Get(apiURL)

	if err == nil && resp.StatusCode == http.StatusOK {
		log.Println("RTMP server is already running.")
		return nil
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RTMP server returned unexpected status: %s", resp.Status)
	}

	log.Printf("RTMP server is running and reachable at port %s.", srs.Port)

	return nil

}

func (srs *SimpleRealtimeServer) Stop() error {

	cmd := exec.Command(srs.SRSPath, "-s", "stop")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to stop RTMP server: %v, output: %s", err, output)
	}

	log.Println("RTMP server stopped")

	return nil

}

func (srs *SimpleRealtimeServer) Reload() error {

	log.Printf("Reloading RTMP server with %d streams configured...", len(srs.Streams))

	// Add timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	cmd := exec.CommandContext(ctx, srs.SRSPath, "-s", "reload")

	output, err := cmd.CombinedOutput()

	if err != nil {

		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("RTMP server reload timed out after 5 seconds")
		}

		return fmt.Errorf("failed to reload RTMP server with new configs: %v, output: %s", err, output)

	}

	return nil

}

func (srs *SimpleRealtimeServer) AddStream(key string, destinations []StreamDestination) error {

	srs.StreamsLock.Lock()

	log.Printf("Adding stream with key: %s and %d destinations", key, len(destinations))

	for i, dest := range destinations {
		log.Printf("Destination %d: URL=%s, StreamKey=%s", i, dest.URL, dest.StreamKey)
	}

	srs.Streams[key] = &Stream{
		Key:          key,
		Destinations: destinations,
	}

	log.Printf("Current streams in map: %d", len(srs.Streams))

	srs.StreamsLock.Unlock()

	return srs.Reload()

}

func (srs *SimpleRealtimeServer) RemoveStream(key string) error {

	srs.StreamsLock.Lock()

	delete(srs.Streams, key)

	srs.StreamsLock.Unlock()

	return srs.Reload()

}

func (srs *SimpleRealtimeServer) GetIngestURL(streamKey string) string {
	//TODO - Make this work for production server and get the info dynamically
	host := "localhost"

	port := srs.Port

	if port[0] == ':' {
		port = port[1:]
	}

	return fmt.Sprintf("rtmp://%s:%s/live/%s", host, port, streamKey)

}

func (srs *SimpleRealtimeServer) GetHLSURL(streamKey string) string {

	return fmt.Sprintf("http://localhost:8081/hls/%s.m3u8", streamKey)

}
