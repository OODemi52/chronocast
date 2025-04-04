// internal/services/stream-platforms/youtube/types.go
package youtube

import (
	"google.golang.org/api/youtube/v3"
)

type StreamDetails struct {
	Broadcast *youtube.LiveBroadcast
	Stream    *youtube.LiveStream
	WatchURL  string
}
