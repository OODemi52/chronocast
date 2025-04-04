package youtube

import (
	"context"
	"time"

	"github.com/OODemi52/chronocast-server/internal/types"
	"google.golang.org/api/youtube/v3"
)

func (s *Service) createYouTubeBroadcast(ctx context.Context, ytService *youtube.Service, options types.StreamOptions) (*youtube.LiveBroadcast, error) {

	scheduledStartTime := options.ScheduleTime

	//FIXME - This check is good, but it also at least needs client side validation as well
	if scheduledStartTime.IsZero() {

		scheduledStartTime = time.Now().Add(1 * time.Minute)

	} else if scheduledStartTime.Before(time.Now()) {

		scheduledStartTime = time.Now().Add(1 * time.Minute)

	}

	broadcast := &youtube.LiveBroadcast{
		Snippet: &youtube.LiveBroadcastSnippet{
			Title:              options.Title,
			Description:        options.Description,
			ScheduledStartTime: scheduledStartTime.Format(time.RFC3339),
		},
		Status: &youtube.LiveBroadcastStatus{
			PrivacyStatus:           options.Privacy,
			SelfDeclaredMadeForKids: false,
		},
		ContentDetails: &youtube.LiveBroadcastContentDetails{
			EnableAutoStart: true,
			EnableAutoStop:  true,
		},
	}

	return ytService.LiveBroadcasts.Insert([]string{"snippet", "status", "contentDetails"}, broadcast).Context(ctx).Do()

}
