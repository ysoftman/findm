package youtube

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"regexp"
	"strings"
)

var (
	channelHandleRE = regexp.MustCompile(`^@[A-Za-z0-9_.-]+$`)

	youtubeHosts = map[string]bool{
		"www.youtube.com":   true,
		"youtube.com":       true,
		"m.youtube.com":     true,
		"music.youtube.com": true,
		"youtu.be":          true,
	}
)

func parseYouTubeURL(s string) *url.URL {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil || !youtubeHosts[u.Host] {
		return nil
	}
	return u
}

// IsYouTubeWatchURL reports whether s is a YouTube watch URL (single video).
func IsYouTubeWatchURL(s string) bool {
	u := parseYouTubeURL(s)
	if u == nil {
		return false
	}
	if u.Host == "youtu.be" {
		return strings.TrimPrefix(u.Path, "/") != ""
	}
	if strings.HasPrefix(u.Path, "/watch") || u.Path == "/shorts" || strings.HasPrefix(u.Path, "/shorts/") {
		if strings.HasPrefix(u.Path, "/shorts/") {
			return strings.TrimPrefix(u.Path, "/shorts/") != ""
		}
		return u.Query().Get("v") != ""
	}
	return false
}

// IsExpandableURL reports whether s is a YouTube playlist or channel URL whose
// videos can be enumerated by FetchVideosFromURL.
func IsExpandableURL(s string) bool {
	u := parseYouTubeURL(s)
	if u == nil {
		return false
	}
	if strings.Contains(u.Path, "/playlist") {
		return true
	}
	if u.Query().Get("list") != "" {
		return true
	}
	if strings.HasPrefix(u.Path, "/@") || strings.HasPrefix(u.Path, "/channel/") {
		return true
	}
	return false
}

// IsChannelHandle reports whether s looks like a bare YouTube @handle (no scheme).
func IsChannelHandle(s string) bool {
	return channelHandleRE.MatchString(strings.TrimSpace(s))
}

// ChannelHandleURL returns the full YouTube URL for a bare @handle.
func ChannelHandleURL(handle string) string {
	return "https://www.youtube.com/" + strings.TrimSpace(handle)
}

// FetchVideosFromURL retrieves the videos referenced by a YouTube playlist or
// channel URL using yt-dlp's flat playlist output. Both /playlist?list=...
// and /@handle (or /channel/UC...) URLs are accepted.
func (c *Client) FetchVideosFromURL(rawURL string, max int) ([]Video, error) {
	if rawURL == "" {
		return nil, fmt.Errorf("empty URL")
	}
	if max <= 0 {
		max = 50
	}
	if max > 200 {
		max = 200
	}

	cmd := exec.Command("yt-dlp",
		"-j",
		"--flat-playlist",
		"--playlist-end", fmt.Sprintf("%d", max),
		rawURL,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start yt-dlp: %w", err)
	}

	var videos []Video
	seen := make(map[string]bool)
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	for scanner.Scan() {
		var r ytdlpResult
		if err := json.Unmarshal(scanner.Bytes(), &r); err != nil {
			continue
		}
		if r.ID == "" || seen[r.ID] {
			continue
		}
		// Skip non-video sub-entries (e.g. nested playlists from a channel root).
		if classifyKind(r) != KindVideo {
			continue
		}
		seen[r.ID] = true

		videoURL := r.WebpageURL
		if videoURL == "" {
			videoURL = fmt.Sprintf("https://www.youtube.com/watch?v=%s", r.ID)
		}

		videos = append(videos, Video{
			ID:        r.ID,
			Title:     normalizeText(r.Title),
			Channel:   normalizeText(r.Channel),
			Duration:  formatDuration(r.Duration),
			ViewCount: r.ViewCount,
			URL:       videoURL,
			Kind:      KindVideo,
		})
	}

	if err := cmd.Wait(); err != nil {
		if len(videos) > 0 {
			return videos, nil
		}
		return nil, fmt.Errorf("yt-dlp fetch failed: %w", err)
	}

	return videos, nil
}
