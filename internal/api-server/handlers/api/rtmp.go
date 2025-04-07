package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/OODemi52/chronocast-server/internal/rtmp-server/auth"
)

type OnPublishRequest struct {
	App    string `json:"app"`
	Stream string `json:"stream"`
	TcUrl  string `json:"tcUrl"`
	Client string `json:"client_id"`
	IP     string `json:"ip"`
	Vhost  string `json:"vhost"`
	Param  string `json:"param"`
}

type OnUnPublishRequest struct {
	App    string `json:"app"`
	Stream string `json:"stream"`
	TcUrl  string `json:"tcUrl"`
	Client string `json:"client_id"`
	IP     string `json:"ip"`
	Vhost  string `json:"vhost"`
	Param  string `json:"param"`
}

//TODO - Consider making a standard request type if it confirmed these two are the same

func RTMPPublishedHandler(w http.ResponseWriter, r *http.Request) {

	var req OnPublishRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to parse JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	log.Printf("RTMP auth request received with data: %+v", req)

	streamKey := req.Stream

	log.Printf("Stream key received: %+v", streamKey)

	if streamKey == "" {
		http.Error(w, "Missing stream key", http.StatusBadRequest)
		return
	}

	if !auth.ValidateStreamKey(streamKey) {
		http.Error(w, "Invalid stream key", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("0"))

}

func RTMPUnPublishedHandler(w http.ResponseWriter, r *http.Request) {

	var req OnUnPublishRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to parse JSON payload: %v", err)
		http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
		return
	}

	log.Printf("RTMP publish done with data: %+v", req)

	w.Header().Set("Content-Type", "text/plain")

	w.WriteHeader(http.StatusOK)

	w.Write([]byte("0"))

}
