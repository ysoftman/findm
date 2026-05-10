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

		title := style.Render(fmt.Sprintf("%s%s%s", prefix, kindBadge(v.Kind), v.Title))
		var info string
		switch v.Kind {
		case youtube.KindPlaylist:
			info = "    " + channelStyle.Render("YouTube playlist") + "  " + durationStyle.Render("Press Enter to view")
		case youtube.KindChannel:
			info = "    " + channelStyle.Render("YouTube channel") + "  " + durationStyle.Render("Press Enter to view")
		default:
			info = fmt.Sprintf("    %s  %s  %s",
				channelStyle.Render(v.Channel),
				durationStyle.Render(v.Duration),
				viewCountStyle.Render(formatViewCount(v.ViewCount)),
			)
		}
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

func kindLabel(k youtube.Kind) string {
	switch k {
	case youtube.KindPlaylist:
		return "playlist"
	case youtube.KindChannel:
		return "channel"
	default:
		return "video"
	}
}

func kindBadge(k youtube.Kind) string {
	switch k {
	case youtube.KindPlaylist:
		return "[PL] "
	case youtube.KindChannel:
		return "[CH] "
	default:
		return ""
	}
}

func nextPlayableIdx(videos []youtube.Video, from, step int) int {
	if step == 0 {
		return -1
	}
	for i := from + step; i >= 0 && i < len(videos); i += step {
		if videos[i].Kind.Playable() {
			return i
		}
	}
	return -1
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
