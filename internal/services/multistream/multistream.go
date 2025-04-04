package multistream

import (
	"fmt"

	rtmpserver "github.com/OODemi52/chronocast-server/internal/rtmp-server"
	"github.com/OODemi52/chronocast-server/internal/services/factory"
	"github.com/OODemi52/chronocast-server/internal/types"
)

type MultiStreamService struct {
	Platforms map[string]types.StreamPlatform
}

type PlatformResult struct {
	Platform        string
	Response        types.StreamResponse
	RTMPDestination rtmpserver.StreamDestination
	Error           error
}

func NewMultiStreamService() (*MultiStreamService, error) {
	//TODO - implement more robust server initilization

	platforms := make(map[string]types.StreamPlatform)

	for _, name := range []string{"youtube"} {

		service, err := factory.GetPlatformService(name)

		if err == nil {
			platforms[name] = service
		}

	}

	if len(platforms) == 0 {
		return nil, fmt.Errorf("no platform services initialized")
	}

	return &MultiStreamService{
		Platforms: platforms,
	}, nil

}

func (m *MultiStreamService) CreateMultiStream(platforms []string, options types.StreamOptions) ([]PlatformResult, error) {

	var results []PlatformResult

	var errors []error

	for _, platformName := range platforms {

		service, exists := m.Platforms[platformName]

		if !exists {
			err := fmt.Errorf("platform %s not initialized", platformName)
			results = append(results, PlatformResult{
				Platform: platformName,
				Error:    err,
			})
			errors = append(errors, err)
			continue
		}

		response, err := service.CreateStream(options)

		result := PlatformResult{
			Platform: platformName,
			Response: response,
		}

		if err != nil {
			result.Error = err
			errors = append(errors, fmt.Errorf("failed on %s: %w", platformName, err))
		} else {
			result.RTMPDestination = getRTMPDestination(platformName, response)
		}

		results = append(results, result)

	}

	if len(errors) == len(platforms) {
		return results, fmt.Errorf("all platforms failed: %v", errors)
	}

	return results, nil

}

func getRTMPDestination(platform string, response types.StreamResponse) rtmpserver.StreamDestination {

	switch platform {

	case "youtube":
		return rtmpserver.StreamDestination{
			URL:       "rtmp://a.rtmp.youtube.com/live2/",
			StreamKey: response.StreamKey,
		}

	case "twitch":
		return rtmpserver.StreamDestination{
			URL:       "rtmp://live.twitch.tv/app/",
			StreamKey: response.StreamKey,
		}

	case "facebook":
		return rtmpserver.StreamDestination{
			URL:       "rtmps://live-api-s.facebook.com:443/rtmp/",
			StreamKey: response.StreamKey,
		}

	default:
		return rtmpserver.StreamDestination{}

	}

}
