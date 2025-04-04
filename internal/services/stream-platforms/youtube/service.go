package youtube

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/OODemi52/chronocast-server/internal/types"
)

type Service struct {
	client *Client
}

func NewService() (*Service, error) {

	client, err := NewClient()

	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube client: %w", err)
	}

	return &Service{
		client: client,
	}, nil

}

func (s *Service) Authenticate(w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	if code == "" {
		return fmt.Errorf("missing authorization code")
	}

	token, err := s.client.Authenticate(r.Context(), code)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	fmt.Printf("Authenticated token: %s\n", token.AccessToken)

	fmt.Fprintf(w, "Successfully authenticated with YouTube")

	return nil
}

func (s *Service) CreateStream(options types.StreamOptions) (types.StreamResponse, error) {
	//FIXME - Get token from file (placeholder - implement real storage)
	tokenData, err := os.ReadFile("token.txt")
	if err != nil {
		return types.StreamResponse{}, fmt.Errorf("failed to read token file: %w", err)
	}
	accessToken := string(tokenData)

	ctx := context.Background()

	ytService, err := s.client.GetYouTubeService(ctx, accessToken)

	if err != nil {
		return types.StreamResponse{}, fmt.Errorf("failed to get YouTube service: %w", err)
	}

	broadcast, err := s.createYouTubeBroadcast(ctx, ytService, options)

	if err != nil {
		return types.StreamResponse{}, fmt.Errorf("failed to create YouTube broadcast: %w", err)
	}

	stream, err := s.createYouTubeStream(ctx, ytService)

	if err != nil {
		return types.StreamResponse{}, fmt.Errorf("failed to create YouTube stream: %w", err)
	}

	err = s.bindYouTubeBroadcastToYouTubeStream(ctx, ytService, broadcast.Id, stream.Id)

	if err != nil {
		return types.StreamResponse{}, fmt.Errorf("failed to bind YouTube broadcast to YouTube stream: %w", err)
	}

	return types.StreamResponse{
		Platform:  "YouTube",
		StreamID:  broadcast.Id,
		URL:       fmt.Sprintf("https://youtube.com/watch?v=%s", broadcast.Id),
		StreamKey: stream.Cdn.IngestionInfo.StreamName,
	}, nil

}

func (s *Service) UpdateStream(id string, options types.StreamOptions) error {
	//TODO - Will implement later
	return fmt.Errorf("update stream not implemented yet")
}

func (s *Service) DeleteStream(id string) error {
	//TODO - Will implement later
	return fmt.Errorf("delete stream not implemented yet")
}
