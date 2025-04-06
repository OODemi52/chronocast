package rtmpserver

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
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
	Port           string
	ConfigPath     string
	SRSPath        string
	SRSProcess     *os.Process
	Streams        map[string]*Stream
	StreamsLock    sync.RWMutex
	ConfigTemplate *template.Template
	ConfigDir      string
}

func NewServer(port string) (*SimpleRealtimeServer, error) {

	workDir, err := os.Getwd()

	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %v", err)
	}

	configDir := filepath.Join(workDir, "configs")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %v", err)
	}

	configPath := filepath.Join(configDir, "nginx.conf")

	if err := os.MkdirAll("/tmp/hls", 0755); err != nil {
		log.Printf("Warning: failed to create HLS directory: %v", err)
	}

	configTemplatePath := filepath.Join(configDir, "nginx.template")

	tmpl, err := template.ParseFiles(configTemplatePath)

	if err != nil {
		return nil, fmt.Errorf("failed to parse config template: %v", err)
	}

	return &SimpleRealtimeServer{
		Port:           port,
		ConfigPath:     configPath,
		SRSPath:        "/opt/homebrew/bin/nginx", // Adjust based on your Nginx installation //TODO - Make dynamic
		Streams:        make(map[string]*Stream),
		ConfigTemplate: tmpl,
	}, nil
}

func (srs *SimpleRealtimeServer) Start() error {

	rtmpPortInUse := exec.Command("lsof", "-i", fmt.Sprintf(":%s", strings.TrimPrefix(srs.Port, ":")))

	rtmpOutput, _ := rtmpPortInUse.CombinedOutput()

	hlsPortInUse := exec.Command("lsof", "-i", ":8081")

	hlsOutput, _ := hlsPortInUse.CombinedOutput()

	if strings.Contains(string(rtmpOutput), "nginx") || strings.Contains(string(hlsOutput), "nginx") {

		log.Println("Nginx is running on our ports, stopping it...")

		stopCmd := exec.Command(srs.SRSPath, "-s", "stop", "-c", srs.ConfigPath)

		if err := stopCmd.Run(); err != nil {
			log.Printf("Warning: failed to stop existing Nginx: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	if err := srs.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	output, err := exec.Command(srs.SRSPath, "-c", srs.ConfigPath).CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to start Nginx RTMP server: %v, output: %s", err, output)
	}

	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("pgrep", "nginx")

	if err := cmd.Run(); err != nil {
		//TODO - implement retries (or look into timeout using context)
		return fmt.Errorf("nginx does not appear to be running after start: %v", err)
	}

	log.Printf("RTMP server started on port %s", srs.Port)

	return nil

}

func (srs *SimpleRealtimeServer) Stop() error {

	cmd := exec.Command(srs.SRSPath, "-s", "stop")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to stop Nginx: %v, output: %s", err, output)
	}

	log.Println("RTMP server stopped")

	return nil

}

func (srs *SimpleRealtimeServer) Reload() error {
	log.Printf("Reloading nginx with %d streams configured...", len(srs.Streams))

	// Add more logging around the generate config call
	log.Println("Starting config generation...")

	if err := srs.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	log.Println("Config generation successful, reading config file...")

	// Debug code
	configContent, err := os.ReadFile(srs.ConfigPath)
	if err == nil {
		log.Printf("Config file contents (length: %d bytes)", len(configContent))
	} else {
		log.Printf("Error reading config file: %v", err)
	}

	log.Println("Executing nginx reload command...")

	// Add timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, srs.SRSPath, "-s", "reload")
	output, err := cmd.CombinedOutput()

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("nginx reload timed out after 5 seconds")
		}
		return fmt.Errorf("failed to reload Nginx: %v, output: %s", err, output)
	}

	log.Printf("Nginx reload command completed successfully")
	log.Printf("Reload output: %s", output)
	log.Println("RTMP server configuration reloaded")

	return nil
}

func (srs *SimpleRealtimeServer) generateConfig() error {

	srs.StreamsLock.RLock()

	defer srs.StreamsLock.RUnlock()

	log.Printf("Generating nginx config with %d streams", len(srs.Streams))

	file, err := os.Create(srs.ConfigPath)

	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	defer file.Close()

	data := struct {
		Port    string
		Streams map[string]*Stream
	}{
		Port:    strings.TrimPrefix(srs.Port, ":"),
		Streams: srs.Streams, //NOTE - Look into possibly populating this with streams from a db in the case of restarting.
	}

	if err := srs.ConfigTemplate.Execute(file, data); err != nil {
		return fmt.Errorf("failed to populate template: %v", err)
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
