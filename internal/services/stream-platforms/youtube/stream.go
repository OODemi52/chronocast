package youtube

import (
	"context"

	"google.golang.org/api/youtube/v3"
)

func (s *Service) createYouTubeStream(ctx context.Context, ytService *youtube.Service) (*youtube.LiveStream, error) {

	stream := &youtube.LiveStream{
		Snippet: &youtube.LiveStreamSnippet{
			Title: "Chronocast Stream", // Replace with passed in option
		},
		Cdn: &youtube.CdnSettings{
			FrameRate:     "variable",
			IngestionType: "rtmp",
			Resolution:    "variable",
		},
	}

	return ytService.LiveStreams.Insert([]string{"snippet", "cdn"}, stream).Context(ctx).Do()

}

func (s *Service) bindYouTubeBroadcastToYouTubeStream(ctx context.Context, ytService *youtube.Service, broadcastId, streamId string) error {

	_, err := ytService.LiveBroadcasts.Bind(broadcastId, []string{"id"}).StreamId(streamId).Context(ctx).Do()

	return err

}
