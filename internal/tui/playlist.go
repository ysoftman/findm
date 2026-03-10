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

	// Each playlist item takes 1 line
	visibleCount := len(names)
	if height > 0 {
		// Subtract 2 lines for title + blank line
		available := height - 2
		if available > 0 {
			visibleCount = available
		}
	}
	if visibleCount > len(names) {
		visibleCount = len(names)
	}

	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	if start+visibleCount > len(names) {
		start = len(names) - visibleCount
	}
	if start < 0 {
		start = 0
	}
	end := start + visibleCount
	if end > len(names) {
		end = len(names)
	}

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

func renderPlaylistDetail(pl *playlist.Playlist, cursor int, height int) string {
	var sb strings.Builder
	sb.WriteString(titleStyle.Render(fmt.Sprintf("Playlist: %s", pl.Name)) + "\n\n")

	if len(pl.Tracks) == 0 {
		sb.WriteString(normalItemStyle.Render("Empty playlist.") + "\n")
		return sb.String()
	}

	// Each track takes 2 lines (title + info)
	itemHeight := 2
	visibleCount := len(pl.Tracks)
	if height > 0 {
		// Subtract 2 lines for title + blank line
		available := height - 2
		if available > 0 {
			visibleCount = available / itemHeight
		}
		if visibleCount < 1 {
			visibleCount = 1
		}
	}
	if visibleCount > len(pl.Tracks) {
		visibleCount = len(pl.Tracks)
	}

	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	if start+visibleCount > len(pl.Tracks) {
		start = len(pl.Tracks) - visibleCount
	}
	if start < 0 {
		start = 0
	}
	end := start + visibleCount
	if end > len(pl.Tracks) {
		end = len(pl.Tracks)
	}

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
		sb.WriteString(title + "\n")
		sb.WriteString(info + "\n")
	}

	if end < len(pl.Tracks) {
		sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(pl.Tracks)-end)) + "\n")
	}

	return sb.String()
}
