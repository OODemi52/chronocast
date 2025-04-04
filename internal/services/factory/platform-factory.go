package factory

import (
	"fmt"

	"github.com/OODemi52/chronocast-server/internal/services/stream-platforms/youtube"
	"github.com/OODemi52/chronocast-server/internal/types"
)

// GetPlatformService creates and returns the appropriate platform service
// based on the provided platform name. It supports different streaming
// platforms by delegating the creation of the service to the corresponding
// implementation.
//
// Parameters:
//   - platform: A string representing the name of the streaming platform
//     (e.g., "youtube").
//
// Returns:
//   - types.StreamPlatform: An interface representing the platform service
//     implementation.
//   - error: An error if the platform is unsupported or if there is an issue
//     creating the service.
//
// Supported Platforms:
//   - "youtube": Returns a YouTube streaming platform service.
//
// Example:
//
//	service, err := GetPlatformService("youtube")
//	if err != nil {
//	    log.Fatalf("Error: %v", err)
//	}
func GetPlatformService(platform string) (types.StreamPlatform, error) {

	switch platform {

	case "youtube":
		service, err := youtube.NewService()

		if err != nil {
			return nil, fmt.Errorf("failed to initialize YouTube service: %w", err)
		}

		return service, nil

	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)

	}

}
