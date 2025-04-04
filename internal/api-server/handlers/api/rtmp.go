package handlers

import (
	"log"
	"net/http"

	"github.com/OODemi52/chronocast-server/internal/rtmp-server/auth"
)

func RTMPAuthHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		log.Printf("RTMP auth parse error: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	log.Printf("RTMP auth request received with data: %+v", r.Form)

	streamKey := r.FormValue("name")

	log.Printf("Stream key received: %+v", streamKey)

	if streamKey == "" {
		http.Error(w, "Missing stream key", http.StatusBadRequest)
		return
	}

	if !auth.ValidateStreamKey(streamKey) {
		http.Error(w, "Invalid stream key", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)

}

func RTMPPublishDoneHandler(w http.ResponseWriter, r *http.Request) {

	if err := r.ParseForm(); err != nil {
		log.Printf("RTMP publish done parse error: %v", err)
		return
	}

	log.Printf("RTMP publish done with data: %+v", r.Form)

	w.WriteHeader(http.StatusOK)

}
