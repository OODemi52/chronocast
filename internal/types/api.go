package types

type CreateStreamRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Destinations []string `json:"destinations"`
}

type StreamAPIResponse struct {
	StreamKey   string `json:"streamKey"`
	IngestURL   string `json:"ingestUrl"`
	HLSPlayURL  string `json:"hlsPlayUrl"`
	Title       string `json:"title"`
	Description string `json:"description"`
}
