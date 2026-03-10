package youtube

import (
	"fmt"
	"os/exec"
)

// Video represents a YouTube video with relevant metadata.
type Video struct {
	ID        string
	Title     string
	Channel   string
	Duration  string
	ViewCount uint64
	URL       string
}

// Client wraps yt-dlp for YouTube search and metadata retrieval.
type Client struct{}

// NewClient creates a new yt-dlp based YouTube client.
// Returns an error if yt-dlp is not installed.
func NewClient() (*Client, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return nil, fmt.Errorf("yt-dlp is not installed. Install it: brew install yt-dlp")
	}
	return &Client{}, nil
}
