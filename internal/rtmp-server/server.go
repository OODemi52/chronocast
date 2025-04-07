package rtmpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
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

	log.Println("Stopping RTMP server via Docker...")

	cmd := exec.Command("docker", "ps", "-q", "--filter", "ancestor=ossrs/srs:latest")

	containerID, err := cmd.Output()

	if err != nil || len(containerID) == 0 {
		return fmt.Errorf("failed to find SRS container: %v", err)
	}

	stopCmd := exec.Command("docker", "stop", strings.TrimSpace(string(containerID)))

	output, err := stopCmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to stop the RTMP server via Docker: %v, output: %s", err, output)
	}

	log.Println("RTMP server stopped successfully.")

	return nil

}

func (srs *SimpleRealtimeServer) Reload() error {

	log.Println("Reloading RTMP server configuration...")

	apiURL := "http://localhost:1985/api/v1/servers"

	req, err := http.NewRequest("PUT", apiURL, strings.NewReader(`{"action":"reload"}`))

	if err != nil {
		return fmt.Errorf("failed to create reload request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to reload RTMP server: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("RTMP server returned unexpected status during reload: %s", resp.Status)
	}

	log.Println("RTMP server configuration reloaded successfully.")

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
