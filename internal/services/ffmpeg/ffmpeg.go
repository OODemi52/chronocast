package ffmpeg

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
)

type FFmpegService struct {
	Processes   map[string]*exec.Cmd
	ProcessLock sync.RWMutex
}

func NewFFmpegService() *FFmpegService {

	return &FFmpegService{
		Processes: make(map[string]*exec.Cmd),
	}

}

func (fs *FFmpegService) StartProcess(streamKey, inputURL, outputURL string) error {

	fs.ProcessLock.RLock()

	_, exists := fs.Processes[streamKey]

	fs.ProcessLock.RUnlock()

	if exists {
		return fmt.Errorf("FFmpeg process already exists for stream key %s", streamKey)
	}

	cmd := exec.Command("ffmpeg",
		"-loglevel", "debug",
		"-i", inputURL,
		"-c:v", "copy",
		"-c:a", "aac",
		"-f", "flv",
		outputURL,
	)

	cmd.Stderr = log.Writer()

	if err := cmd.Start(); err != nil {
		log.Printf("Failed to start FFmpeg process for stream key %s: %v", streamKey, err)
		return err
	}

	fs.ProcessLock.Lock()

	fs.Processes[streamKey] = cmd

	fs.ProcessLock.Unlock()

	log.Printf("Started FFmpeg process for stream key %s, output: %s", streamKey, outputURL)

	return nil

}

func (fs *FFmpegService) StopProcess(streamKey string) error {

	fs.ProcessLock.Lock()

	cmd, exists := fs.Processes[streamKey]

	if exists {
		delete(fs.Processes, streamKey)
	}

	fs.ProcessLock.Unlock()

	if !exists {
		return nil
	}

	if err := cmd.Process.Kill(); err != nil {
		return err
	}

	log.Printf("Stopped FFmpeg process for stream key %s", streamKey)

	return nil

}

func (fs *FFmpegService) StopAllProcesses() {

	fs.ProcessLock.Lock()

	defer fs.ProcessLock.Unlock()

	for key, cmd := range fs.Processes {

		if err := cmd.Process.Kill(); err != nil {
			log.Printf("Failed to stop FFmpeg process for key %s: %v", key, err)
		} else {
			log.Printf("Stopped FFmpeg process for key %s", key)
		}

		delete(fs.Processes, key)

	}

}
