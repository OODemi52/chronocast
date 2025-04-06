package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
	"github.com/OODemi52/chronocast-server/internal/rtmp-server/auth"
	"github.com/OODemi52/chronocast-server/internal/services/multistream"
	"github.com/OODemi52/chronocast-server/internal/types"
)

func GenerateStreamKeyHandler(w http.ResponseWriter, r *http.Request) {
	//FIXME - Instead of passing a user id, use JWTs with an access/refresh token scheme
	//		  and decode to get the user id. Short lived access, long lived refresh in http cookie
	//		  Add CRSF proctection.

	var request struct {
		UserID string `json:"userID"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.UserID == "" {
		http.Error(w, "Missing userID", http.StatusBadRequest)
		return
	}

	streamKey, err := auth.GenerateStreamKey(request.UserID)

	if err != nil {
		http.Error(w, "Failed to generate stream key", http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"streamKey": streamKey,
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(response)

}

func CreateStreamHandler(rtmpServer *rtmpserver.SimpleRealtimeServer) http.HandlerFunc {
	//TODO - This function is handling to many different responsibilities
	//       Need to reasses scope and split it up

	return func(w http.ResponseWriter, r *http.Request) {

		multiStreamService, err := multistream.NewMultiStreamService()

		if err != nil {
			log.Printf("Warning: Failed to initialize MultiStreamService: %v", err)
		}

		if r.Method != http.MethodPost {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var request struct {
			UserID       string   `json:"userID"`
			Title        string   `json:"title"`
			Description  string   `json:"description"`
			Destinations []string `json:"destinations"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		log.Printf("Received CreateStream request: UserID=%s, Title=%s, Description=%s, Destinations=%v",
			request.UserID, request.Title, request.Description, request.Destinations)

		if request.UserID == "" {
			http.Error(w, "Missing userID", http.StatusBadRequest)
			return
		} //TODO - Added checks for other fields in request type

		streamKey, exists := auth.GetStreamKeyForUser(request.UserID)

		if !exists {
			http.Error(w, "Stream key not found for user", http.StatusUnauthorized)
			return
		}

		var destinations []rtmpserver.StreamDestination

		if multiStreamService != nil {

			//TODO - Form validation and sanitization (client and server)
			// Convert destinations to lowercase for case-insensitive matching
			for i, destination := range request.Destinations {
				request.Destinations[i] = strings.ToLower(destination)
			}

			results, err := multiStreamService.CreateMultiStream(
				request.Destinations,
				types.StreamOptions{
					Title:       request.Title,
					Description: request.Description,
					Privacy:     "public", //FIXME - change to be set by user instead of defaulting to public
				},
			)

			if err != nil {
				log.Printf("Warning: Some stream platforms failed: %v", err)
			}

			for _, result := range results {

				if result.Error == nil {

					log.Printf("Adding destination: %s", result.RTMPDestination.URL)
					destinations = append(destinations, result.RTMPDestination)

				} else {

					log.Printf("Failed to create stream on %s: %v", result.Platform, result.Error)

				}
			}

		} else {

			errMsg := "Multi-streaming service temporarily unavailable"

			log.Printf("ERROR: MultiStreamService unavailable: %v", err)

			http.Error(w, errMsg, http.StatusServiceUnavailable)

			return

		}

		if err := rtmpServer.AddStream(streamKey, destinations); err != nil {
			http.Error(w, "Failed to create stream", http.StatusInternalServerError)
			return
		}

		response := struct {
			StreamKey   string `json:"streamKey"`
			IngestURL   string `json:"ingestUrl"`
			HLSPlayURL  string `json:"hlsPlayUrl"`
			Title       string `json:"title"`
			Description string `json:"description"`
		}{
			StreamKey:   streamKey,
			IngestURL:   rtmpServer.GetIngestURL(streamKey),
			HLSPlayURL:  rtmpServer.GetHLSURL(streamKey),
			Title:       request.Title,
			Description: request.Description,
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(response)

	}
}

func ManageStreamHandler(rtmpServer *rtmpserver.SimpleRealtimeServer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract stream key from URL path
		parts := strings.Split(r.URL.Path, "/")
		if len(parts) < 3 {
			http.Error(w, "Invalid stream ID", http.StatusBadRequest)
			return
		}
		streamKey := parts[len(parts)-1]

		switch r.Method {
		case http.MethodDelete:
			// Revoke the stream key
			auth.RevokeStreamKey(streamKey)

			// Remove the stream from the RTMP server
			if err := rtmpServer.RemoveStream(streamKey); err != nil {
				http.Error(w, "Failed to delete stream", http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
