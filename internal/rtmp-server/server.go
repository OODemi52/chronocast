package rtmpserver

import (
	"encoding/json"
	"fmt"
	"io"
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

	if key == "" {
		return fmt.Errorf("stream key cannot be empty")
	}

	if len(destinations) == 0 {
		return fmt.Errorf("destinations cannot be empty")
	}

	log.Printf("Adding stream with key: %s and %d destinations...", key, len(destinations))

	for i, dest := range destinations {
		log.Printf("Destination %d: URL=%s, StreamKey=%s", i, dest.URL, dest.StreamKey)
	}

	payload := map[string]any{
		"id":           key,
		"destinations": destinations,
	}

	jsonPayload, err := json.Marshal(payload)

	if err != nil {
		return fmt.Errorf("failed to convert payload to json: %v", err)
	}

	apiURL := "http://localhost:1985/api/v1/streams"

	req, err := http.NewRequest("POST", apiURL, strings.NewReader(string(jsonPayload)))

	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to send HTTP request to SRS API: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RTMP server returned unexpected status: %s, body: %s", resp.Status, string(body))
	}

	log.Println("Stream added successfully added.")

	srs.StreamsLock.Lock()

	defer srs.StreamsLock.Unlock()

	srs.Streams[key] = &Stream{
		Key:          key,
		Destinations: destinations,
	}

	return nil

}

func (srs *SimpleRealtimeServer) RemoveStream(key string) error {

	if key == "" {
		return fmt.Errorf("stream key cannot be empty")
	}

	log.Printf("Removing stream with key: %s ...", key)

	apiURL := fmt.Sprintf("http://localhost:1985/api/v1/streams/%s", key)

	req, err := http.NewRequest("DELETE", apiURL, nil)

	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		return fmt.Errorf("failed to send HTTP request to SRS API: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RTMP server returned unexpected status: %s, body: %s", resp.Status, string(body))
	}

	log.Println("Stream removed successfully.")

	srs.StreamsLock.Lock()

	defer srs.StreamsLock.Unlock()

	delete(srs.Streams, key)

	return nil

}

func (srs *SimpleRealtimeServer) GetIngestURL(streamKey string) string {

	if streamKey == "" {
		log.Println("Stream key is empty. Returning an empty ingest URL.")
		return ""
	}
	//TODO - Make this work for production server and get the info dynamically
	host := "localhost"

	port := srs.Port

	if port[0] == ':' {
		port = port[1:]
	}

	return fmt.Sprintf("rtmp://%s:%s/live/%s", host, port, streamKey)

}

func (srs *SimpleRealtimeServer) GetHLSURL(streamKey string) string {

	if streamKey == "" {
		log.Println("Stream key is empty. Returning an empty HLS URL.")
		return ""
	}
	//TODO - Make this work for production server and get the info dynamically
	host := "localhost"

	return fmt.Sprintf("http://%s:8080/hls/%s.m3u8", host, streamKey)

}
