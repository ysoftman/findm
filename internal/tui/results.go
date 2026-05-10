package tui

import (
	"fmt"
	"strings"

	"github.com/ysoftman/findm/internal/youtube"
)

func renderResults(videos []youtube.Video, cursor int, height int, canLoadMore bool) string {
	if len(videos) == 0 {
		return normalItemStyle.Render("No results found.")
	}

	totalItems := len(videos)
	if canLoadMore {
		totalItems++
	}

	visibleCount := fittedVisibleCount(totalItems, cursor, height, func(start, end int) int {
		lines := 0
		for i := start; i < end; i++ {
			if i < len(videos) {
				lines += 3
			} else {
				lines++
			}
			if i < end-1 {
				lines++
			}
		}
		if start > 0 {
			lines++
		}
		if end < totalItems {
			lines += 2
		}
		return lines
	})
	start, end := viewportBounds(totalItems, cursor, visibleCount)

	var sb strings.Builder

	if start > 0 {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
	}

	for i := start; i < end; i++ {
		if i == len(videos) {
			prefix := "  "
			style := normalItemStyle
			if i == cursor {
				prefix = "▸ "
				style = selectedItemStyle
			}
			sb.WriteString(style.Render(prefix + "Load more results..."))
			continue
		}

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
		url := "    " + urlStyle.Render(videoURL(v))

		sb.WriteString(title + "\n")
		sb.WriteString(info + "\n")
		sb.WriteString(url + "\n")
		if i < end-1 {
			sb.WriteString("\n")
		}
	}

	if end < totalItems {
		sb.WriteString("\n" + scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", totalItems-end)))
	}

	return sb.String()
}

func videoURL(v youtube.Video) string {
	if v.URL != "" {
		return v.URL
	}
	if v.ID != "" {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", v.ID)
	}
	return ""
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
