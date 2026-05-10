package tui

import (
	"fmt"
	"strings"

	"github.com/ysoftman/findm/internal/playlist"
)

func renderPlaylistList(names []string, cursor int, height int) string {
	if len(names) == 0 {
		return normalItemStyle.Render("No playlists yet. Press 'c' to create one.")
	}

	var sb strings.Builder
	sb.WriteString(titleStyle.Render("Playlists") + "\n\n")

	visibleCount := fittedVisibleCount(len(names), cursor, height, func(start, end int) int {
		lines := 2 + end - start
		if start > 0 {
			lines++
		}
		if end < len(names) {
			lines++
		}
		return lines
	})
	start, end := viewportBounds(len(names), cursor, visibleCount)

	if start > 0 {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
	}

	for i := start; i < end; i++ {
		name := names[i]
		prefix := "  "
		style := normalItemStyle
		if i == cursor {
			prefix = "▸ "
			style = selectedItemStyle
		}
		sb.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, name)) + "\n")
	}

	if end < len(names) {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(names)-end)) + "\n")
	}

	return sb.String()
}

func trackURL(t playlist.Track) string {
	if t.URL != "" {
		return t.URL
	}
	if t.VideoID != "" {
		return fmt.Sprintf("https://www.youtube.com/watch?v=%s", t.VideoID)
	}
	return ""
}

func renderPlaylistDetail(pl *playlist.Playlist, cursor int, height int) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(fmt.Sprintf("Playlist: %s", pl.Name)) + "\n\n")

	if len(pl.Tracks) == 0 {
		sb.WriteString(normalItemStyle.Render("Empty playlist.") + "\n")
		return sb.String()
	}

	visibleCount := fittedVisibleCount(len(pl.Tracks), cursor, height, func(start, end int) int {
		lines := 2 + (end-start)*3
		if start > 0 {
			lines++
		}
		if end < len(pl.Tracks) {
			lines++
		}
		return lines
	})
	start, end := viewportBounds(len(pl.Tracks), cursor, visibleCount)

	if start > 0 {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
	}

	for i := start; i < end; i++ {
		t := pl.Tracks[i]
		prefix := "  "
		style := normalItemStyle
		if i == cursor {
			prefix = "▸ "
			style = selectedItemStyle
		}
		title := style.Render(fmt.Sprintf("%s%s", prefix, t.Title))
		info := fmt.Sprintf("    %s", channelStyle.Render(t.Channel))
		url := "    " + urlStyle.Render(trackURL(t))
		sb.WriteString(title + "\n")
		sb.WriteString(info + "\n")
		sb.WriteString(url + "\n")
	}

	if end < len(pl.Tracks) {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(pl.Tracks)-end)) + "\n")
	}

	return sb.String()
}
