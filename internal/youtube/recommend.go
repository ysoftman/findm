package youtube

import (
	"fmt"
	"regexp"
	"strings"
)

var noisePattern = regexp.MustCompile(`(?i)(\(|\[)(official\s*(video|mv|music\s*video|audio)|m/?v|lyrics|audio)(\)|\])`)

// Recommend finds similar music based on a video's title and channel.
// It fetches the video metadata via yt-dlp, then performs a related search.
func (c *Client) Recommend(videoID string, maxResults int64) ([]Video, error) {
	if maxResults <= 0 {
		maxResults = 10
	}
	if maxResults > 50 {
		maxResults = 50
	}

	// Get the video metadata to extract keywords
	meta, err := c.fetchVideoMetadata(videoID)
	if err != nil {
		return nil, fmt.Errorf("failed to get video details: %w", err)
	}

	if meta.Title == "" {
		return nil, fmt.Errorf("video not found: %s", videoID)
	}

	// Build a search query from the video's channel and cleaned title
	query := buildRecommendQuery(meta.Title, meta.Channel)

	// Search for similar videos, +1 to exclude the original
	results, err := c.Search(query, maxResults+1, 0)
	if err != nil {
		return nil, fmt.Errorf("recommend search failed: %w", err)
	}

	// Exclude the original video
	filtered := make([]Video, 0, len(results))
	for _, v := range results {
		if v.ID != videoID {
			filtered = append(filtered, v)
		}
	}

	// Limit to maxResults
	if int64(len(filtered)) > maxResults {
		filtered = filtered[:maxResults]
	}

	return filtered, nil
}

// buildRecommendQuery creates a search query from video metadata.
func buildRecommendQuery(title, channel string) string {
	cleanTitle := strings.TrimSpace(noisePattern.ReplaceAllString(title, ""))

	// Use channel name + cleaned title as query for better recommendations
	return fmt.Sprintf("%s %s", channel, cleanTitle)
}
