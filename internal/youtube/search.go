package youtube

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"strings"
)

// ytdlpResult represents a single JSON line from yt-dlp output.
type ytdlpResult struct {
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Channel    string  `json:"channel"`
	Duration   float64 `json:"duration"`
	ViewCount  uint64  `json:"view_count"`
	WebpageURL string  `json:"webpage_url"`
	URL        string  `json:"url"`
}

// Search performs a keyword search for music videos using yt-dlp.
func (c *Client) Search(query string, maxResults int64) ([]Video, error) {
	if maxResults <= 0 {
		maxResults = 10
	}

	searchQuery := fmt.Sprintf("ytsearch%d:%s", maxResults, query)
	cmd := exec.Command("yt-dlp", "-j", "--flat-playlist", searchQuery)

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
	// Increase buffer size for long JSON lines
	scanner.Buffer(make([]byte, 0, 256*1024), 256*1024)

	for scanner.Scan() {
		var result ytdlpResult
		if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
			continue
		}
		if result.ID == "" || seen[result.ID] {
			continue
		}
		seen[result.ID] = true

		url := result.WebpageURL
		if url == "" {
			url = fmt.Sprintf("https://www.youtube.com/watch?v=%s", result.ID)
		}

		videos = append(videos, Video{
			ID:        result.ID,
			Title:     result.Title,
			Channel:   result.Channel,
			Duration:  formatDuration(result.Duration),
			ViewCount: result.ViewCount,
			URL:       url,
		})
	}

	if err := cmd.Wait(); err != nil {
		// If we got some results, return them despite exit error
		if len(videos) > 0 {
			return videos, nil
		}
		return nil, fmt.Errorf("yt-dlp search failed: %w", err)
	}

	return videos, nil
}

// formatDuration converts seconds (float64) to "M:SS" or "H:MM:SS" format.
func formatDuration(seconds float64) string {
	if seconds <= 0 {
		return "0:00"
	}

	total := int(math.Round(seconds))
	h := total / 3600
	m := (total % 3600) / 60
	s := total % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// fetchVideoMetadata retrieves full metadata for a single video using yt-dlp.
func (c *Client) fetchVideoMetadata(videoID string) (*ytdlpResult, error) {
	url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	cmd := exec.Command("yt-dlp", "-j", "--no-playlist", url)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video metadata: %w", err)
	}

	var result ytdlpResult
	if err := json.Unmarshal([]byte(strings.TrimSpace(string(output))), &result); err != nil {
		return nil, fmt.Errorf("failed to parse video metadata: %w", err)
	}

	return &result, nil
}
