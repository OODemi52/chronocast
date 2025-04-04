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

type NginxServer struct {
	Port           string
	ConfigPath     string
	NginxPath      string
	NginxProcess   *os.Process
	Streams        map[string]*Stream
	StreamsLock    sync.RWMutex
	ConfigTemplate *template.Template
	ConfigDir      string
}

func NewServer(port string) (*NginxServer, error) {

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

	return &NginxServer{
		Port:           port,
		ConfigPath:     configPath,
		NginxPath:      "/opt/homebrew/bin/nginx", // Adjust based on your Nginx installation //TODO - Make dynamic
		Streams:        make(map[string]*Stream),
		ConfigTemplate: tmpl,
	}, nil
}

func (ns *NginxServer) Start() error {

	rtmpPortInUse := exec.Command("lsof", "-i", fmt.Sprintf(":%s", strings.TrimPrefix(ns.Port, ":")))

	rtmpOutput, _ := rtmpPortInUse.CombinedOutput()

	hlsPortInUse := exec.Command("lsof", "-i", ":8081")

	hlsOutput, _ := hlsPortInUse.CombinedOutput()

	if strings.Contains(string(rtmpOutput), "nginx") || strings.Contains(string(hlsOutput), "nginx") {

		log.Println("Nginx is running on our ports, stopping it...")

		stopCmd := exec.Command(ns.NginxPath, "-s", "stop", "-c", ns.ConfigPath)

		if err := stopCmd.Run(); err != nil {
			log.Printf("Warning: failed to stop existing Nginx: %v", err)
		}

		time.Sleep(1 * time.Second)
	}

	if err := ns.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	output, err := exec.Command(ns.NginxPath, "-c", ns.ConfigPath).CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to start Nginx RTMP server: %v, output: %s", err, output)
	}

	time.Sleep(500 * time.Millisecond)

	cmd := exec.Command("pgrep", "nginx")

	if err := cmd.Run(); err != nil {
		//TODO - implement retries (or look into timeout using context)
		return fmt.Errorf("nginx does not appear to be running after start: %v", err)
	}

	log.Printf("RTMP server started on port %s", ns.Port)

	return nil

}

func (ns *NginxServer) Stop() error {

	cmd := exec.Command(ns.NginxPath, "-s", "stop")

	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("failed to stop Nginx: %v, output: %s", err, output)
	}

	log.Println("RTMP server stopped")

	return nil

}

func (ns *NginxServer) Reload() error {
	log.Printf("Reloading nginx with %d streams configured...", len(ns.Streams))

	// Add more logging around the generate config call
	log.Println("Starting config generation...")

	if err := ns.generateConfig(); err != nil {
		return fmt.Errorf("failed to generate config: %v", err)
	}

	log.Println("Config generation successful, reading config file...")

	// Debug code
	configContent, err := os.ReadFile(ns.ConfigPath)
	if err == nil {
		log.Printf("Config file contents (length: %d bytes)", len(configContent))
	} else {
		log.Printf("Error reading config file: %v", err)
	}

	log.Println("Executing nginx reload command...")

	// Add timeout context to prevent hanging
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, ns.NginxPath, "-s", "reload")
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

func (ns *NginxServer) generateConfig() error {

	ns.StreamsLock.RLock()

	defer ns.StreamsLock.RUnlock()

	log.Printf("Generating nginx config with %d streams", len(ns.Streams))

	file, err := os.Create(ns.ConfigPath)

	if err != nil {
		return fmt.Errorf("failed to create config file: %v", err)
	}

	defer file.Close()

	data := struct {
		Port    string
		Streams map[string]*Stream
	}{
		Port:    strings.TrimPrefix(ns.Port, ":"),
		Streams: ns.Streams, //NOTE - Look into possibly populating this with streams from a db in the case of restarting.
	}

	if err := ns.ConfigTemplate.Execute(file, data); err != nil {
		return fmt.Errorf("failed to populate template: %v", err)
	}

	return nil
}

func (ns *NginxServer) AddStream(key string, destinations []StreamDestination) error {

	ns.StreamsLock.Lock()

	log.Printf("Adding stream with key: %s and %d destinations", key, len(destinations))

	for i, dest := range destinations {
		log.Printf("Destination %d: URL=%s, StreamKey=%s", i, dest.URL, dest.StreamKey)
	}

	ns.Streams[key] = &Stream{
		Key:          key,
		Destinations: destinations,
	}

	log.Printf("Current streams in map: %d", len(ns.Streams))

	ns.StreamsLock.Unlock()

	return ns.Reload()

}

func (ns *NginxServer) RemoveStream(key string) error {

	ns.StreamsLock.Lock()

	delete(ns.Streams, key)

	ns.StreamsLock.Unlock()

	return ns.Reload()

}

func (ns *NginxServer) GetIngestURL(streamKey string) string {
	//TODO - Make this work for production server and get the info dynamically
	host := "localhost"

	port := ns.Port

	if port[0] == ':' {
		port = port[1:]
	}

	return fmt.Sprintf("rtmp://%s:%s/live/%s", host, port, streamKey)

}

func (ns *NginxServer) GetHLSURL(streamKey string) string {

	return fmt.Sprintf("http://localhost:8081/hls/%s.m3u8", streamKey)

}
