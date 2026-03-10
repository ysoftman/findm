package tui

import (
	"fmt"
	"strings"

	"github.com/ysoftman/findm/internal/youtube"
)

func renderResults(videos []youtube.Video, cursor int, height int) string {
	if len(videos) == 0 {
		return normalItemStyle.Render("No results found.")
	}

	// Each item takes 3 lines (title + info + blank line)
	itemHeight := 3
	visibleCount := len(videos)
	if height > 0 {
		visibleCount = height / itemHeight
		if visibleCount < 1 {
			visibleCount = 1
		}
	}
	if visibleCount > len(videos) {
		visibleCount = len(videos)
	}

	// Calculate viewport start based on cursor position
	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	if start+visibleCount > len(videos) {
		start = len(videos) - visibleCount
	}
	if start < 0 {
		start = 0
	}
	end := start + visibleCount
	if end > len(videos) {
		end = len(videos)
	}

	var sb strings.Builder

	if start > 0 {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
	}

	for i := start; i < end; i++ {
		v := videos[i]
		prefix := "  "
		style := normalItemStyle
		if i == cursor {
			prefix = "▸ "
			style = selectedItemStyle
		}

		title := style.Render(fmt.Sprintf("%s%s", prefix, v.Title))
		info := fmt.Sprintf("    %s  %s  %s",
			channelStyle.Render(v.Channel),
			durationStyle.Render(v.Duration),
			viewCountStyle.Render(formatViewCount(v.ViewCount)),
		)

		sb.WriteString(title + "\n")
		sb.WriteString(info + "\n")
		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	if end < len(videos) {
		sb.WriteString("\n" + scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(videos)-end)))
	}

	return sb.String()
}

func formatViewCount(count uint64) string {
	switch {
	case count >= 1_000_000_000:
		return fmt.Sprintf("%.1fB views", float64(count)/1_000_000_000)
	case count >= 1_000_000:
		return fmt.Sprintf("%.1fM views", float64(count)/1_000_000)
	case count >= 1_000:
		return fmt.Sprintf("%.1fK views", float64(count)/1_000)
	default:
		return fmt.Sprintf("%d views", count)
	}
}
