package tui

import (
	"fmt"
	"strings"

	"github.com/ysoftman/findm/internal/player"
)

func renderPlayerBar(p *player.Player, width int) string {
	if width <= 0 {
		width = 80
	}
	style := playerBarStyle.Width(width)

	if p.GetState() == player.Stopped {
		return style.Render("  No track playing  |  /: search  q: quit")
	}

	title := p.CurrentTitle()
	pos := p.GetPosition()
	dur := p.GetDuration()
	vol := p.GetVolume()

	// Truncate title based on available width
	// Layout: "  ▶ Title  [progress] pos/dur  Vol:XX%"
	// Reserve ~45 chars for controls/progress minimum
	maxTitle := width - 45
	if maxTitle < 15 {
		maxTitle = 15
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-3] + "..."
	}

	status := p.StateString()

	// Build progress bar
	progressBar := renderProgressBar(pos, dur, width)

	posStr := formatSeconds(pos)
	durStr := formatSeconds(dur)

	var line string
	if p.GetState() == player.Paused {
		line = fmt.Sprintf("  %s  %s  %s %s/%s  Vol:%d%%", status, title, progressBar, posStr, durStr, vol)
	} else {
		line = fmt.Sprintf("  %s  %s  %s %s/%s  Vol:%d%%", status, title, progressBar, posStr, durStr, vol)
	}

	return style.Render(line)
}

func renderProgressBar(pos, dur float64, width int) string {
	// Progress bar width: scale with terminal width
	barWidth := width / 5
	if barWidth < 8 {
		barWidth = 8
	}
	if barWidth > 30 {
		barWidth = 30
	}

	filled := 0
	if dur > 0 {
		ratio := pos / dur
		if ratio > 1 {
			ratio = 1
		}
		filled = int(ratio * float64(barWidth))
	}
	empty := barWidth - filled

	bar := progressFilledStyle.Render(strings.Repeat("█", filled)) +
		progressEmptyStyle.Render(strings.Repeat("░", empty))

	return "[" + bar + "]"
}

func formatSeconds(s float64) string {
	if s < 0 {
		s = 0
	}
	total := int(s)
	h := total / 3600
	m := (total % 3600) / 60
	sec := total % 60

	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, sec)
	}
	return fmt.Sprintf("%d:%02d", m, sec)
}
