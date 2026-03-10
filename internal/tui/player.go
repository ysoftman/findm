package tui

import (
	"fmt"
	"strings"

	"github.com/ysoftman/findm/internal/player"
	"github.com/ysoftman/findm/internal/visualizer"
)

var playingIcons = []string{"♪ Playing", "♫ Playing", "♬ Playing", "♫ Playing"}
var preparingIcons = []string{"⏳ Preparing.", "⏳ Preparing..", "⏳ Preparing..."}

func animatedStateString(state player.State, frame int) string {
	switch state {
	case player.Preparing:
		return preparingIcons[frame%len(preparingIcons)]
	case player.Playing:
		return playingIcons[frame%len(playingIcons)]
	case player.Paused:
		return "⏸ Paused"
	default:
		return "⏹ Stopped"
	}
}

func renderPlayerBar(p *player.Player, width, frame int) string {
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
	maxTitle := width - 50
	if maxTitle < 15 {
		maxTitle = 15
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-3] + "..."
	}

	status := animatedStateString(p.GetState(), frame)
	progressBar := renderProgressBar(pos, dur, width)
	posStr := formatSeconds(pos)
	durStr := formatSeconds(dur)

	line := fmt.Sprintf("  %s  %s  %s %s/%s  Vol:%d%%", status, title, progressBar, posStr, durStr, vol)
	return style.Render(line)
}

func renderVisualizer(viz *visualizer.Visualizer, width int) string {
	if viz == nil || !viz.IsRunning() {
		return ""
	}

	values := viz.Values()
	peaks := viz.Peaks()
	if len(values) == 0 {
		return ""
	}

	blocks := []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	var sb strings.Builder
	sb.WriteString("  ")
	for i, v := range values {
		idx := int(v * float64(len(blocks)-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= len(blocks) {
			idx = len(blocks) - 1
		}

		ch := string(blocks[idx])

		// Show peak marker if peak is significantly above current value
		isPeak := false
		if i < len(peaks) && peaks[i]-v > 0.15 && peaks[i] > 0.3 {
			peakIdx := int(peaks[i] * float64(len(blocks)-1))
			if peakIdx >= len(blocks) {
				peakIdx = len(blocks) - 1
			}
			if peakIdx > idx {
				ch = string(blocks[peakIdx])
				isPeak = true
			}
		}

		// Color from gradient based on value (0.0-1.0 → 8 color levels)
		var rendered string
		if isPeak {
			rendered = vizPeakStyle.Render(ch)
		} else {
			ci := int(v * float64(len(vizGradient)-1))
			if ci < 0 {
				ci = 0
			}
			if ci >= len(vizGradient) {
				ci = len(vizGradient) - 1
			}
			rendered = vizGradient[ci].Render(ch)
		}

		sb.WriteString(rendered)
	}

	return sb.String()
}

func renderProgressBar(pos, dur float64, width int) string {
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
