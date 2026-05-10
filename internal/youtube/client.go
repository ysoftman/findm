package youtube

import (
	"fmt"
	"os/exec"

	"golang.org/x/text/unicode/norm"
)

// Kind identifies the kind of a search result entry.
type Kind int

const (
	KindVideo Kind = iota
	KindPlaylist
	KindChannel
)

func (k Kind) Playable() bool { return k == KindVideo }

// Video represents a YouTube search result entry.
// For KindPlaylist / KindChannel, Duration and ViewCount may be empty/zero.
type Video struct {
	ID        string
	Title     string
	Channel   string
	Duration  string
	ViewCount uint64
	URL       string
	Kind      Kind
}

// normalizeText returns the NFC-normalized form of s.
// yt-dlp sometimes returns Korean (and other CJK) text in NFD,
// which renders as separated jamo in many terminal fonts.
func normalizeText(s string) string {
	return norm.NFC.String(s)
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
