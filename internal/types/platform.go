package types

import (
	"net/http"
	"time"
)

type StreamOptions struct {
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	Privacy      string    `json:"privacy"`
	ScheduleTime time.Time `json:"scheduleTime"`
}

type StreamResponse struct {
	Platform  string `json:"platform,omitempty"`
	StreamID  string `json:"stream_id,omitempty"`
	URL       string `json:"url,omitempty"`
	StreamKey string `json:"stream_key,omitempty"`
}

type StreamPlatform interface {
	Authenticate(w http.ResponseWriter, r *http.Request) error
	CreateStream(options StreamOptions) (StreamResponse, error)
	UpdateStream(id string, options StreamOptions) error
	DeleteStream(id string) error
}
