package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/ysoftman/findm/internal/player"
	"github.com/ysoftman/findm/internal/visualizer"
)

var playingIcons = []string{"♪ Playing", "♫ Playing", "♬ Playing", "♫ Playing"}
var preparingIcons = []string{"⏳ Preparing.", "⏳ Preparing..", "⏳ Preparing..."}
var loadingDots = []string{".", "..", "..."}

const preparingPlaybackMsg = "Preparing playback"

func animatedLoadingMessage(base string, frame int) string {
	slow := frame / 5
	return base + loadingDots[slow%len(loadingDots)]
}

const (
	visualizerHeight     = 4
	visualizerBarWidth   = 2
	visualizerBarSpacing = 1
)

var visualizerBlocks = []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

func animatedStateString(state player.State, frame int) string {
	// Slow down icon animation (~1s per step at 100ms tick)
	slow := frame / 10
	switch state {
	case player.Preparing:
		return preparingIcons[slow%len(preparingIcons)]
	case player.Playing:
		return playingIcons[slow%len(playingIcons)]
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

	// Truncate title based on available display columns (wide-char aware).
	maxTitle := width - 50
	if maxTitle < 15 {
		maxTitle = 15
	}
	title = runewidth.Truncate(title, maxTitle, "...")

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
	if len(values) == 0 {
		return ""
	}

	var sb strings.Builder

	prefix := ""
	availableWidth := width - len(prefix)
	if availableWidth <= 0 {
		return ""
	}

	barCount := visualizerBarCount(len(values), availableWidth)
	if barCount <= 0 {
		return ""
	}

	for row := visualizerHeight; row >= 1; row-- {
		if row != visualizerHeight {
			sb.WriteByte('\n')
		}
		sb.WriteString(prefix)

		for bar := 0; bar < barCount; bar++ {
			barWidth := visualizerBarWidth

			value := interpolatedVisualizerValue(values, bar, barCount)
			ch := visualizerRowChar(value, row)

			if ch == " " {
				sb.WriteString(strings.Repeat(ch, barWidth))
			} else {
				sb.WriteString(renderVisualizerCell(strings.Repeat(ch, barWidth), row))
			}

			if bar < barCount-1 {
				sb.WriteString(strings.Repeat(" ", visualizerBarSpacing))
			}
		}
	}

	return sb.String()
}

func visualizerBarCount(valueCount, width int) int {
	if valueCount <= 0 || width <= 0 {
		return 0
	}

	count := (width + visualizerBarSpacing) / (visualizerBarWidth + visualizerBarSpacing)
	if count < 1 {
		count = 1
	}
	if count > valueCount {
		count = valueCount
	}
	return count
}

func interpolatedVisualizerValue(values []float64, col, width int) float64 {
	if len(values) == 0 || width <= 0 {
		return 0
	}
	if len(values) == 1 || width == 1 {
		return values[0]
	}

	position := float64(col) * float64(len(values)-1) / float64(width-1)
	left := int(math.Floor(position))
	right := left + 1
	if right >= len(values) {
		return values[left]
	}

	blend := position - float64(left)
	return values[left]*(1-blend) + values[right]*blend
}

func visualizerRowChar(value float64, row int) string {
	rowBottom := float64(row-1) / float64(visualizerHeight)
	rowTop := float64(row) / float64(visualizerHeight)

	if value >= rowTop {
		return "█"
	}
	if value > rowBottom {
		fill := (value - rowBottom) / (rowTop - rowBottom)
		idx := int(math.Ceil(fill*float64(len(visualizerBlocks)))) - 1
		if idx < 0 {
			idx = 0
		}
		if idx >= len(visualizerBlocks) {
			idx = len(visualizerBlocks) - 1
		}
		return string(visualizerBlocks[idx])
	}

	return " "
}

func renderVisualizerCell(ch string, row int) string {
	styleIdx := row - 1
	if styleIdx < 0 {
		styleIdx = 0
	}
	if styleIdx >= len(vizRowStyles) {
		styleIdx = len(vizRowStyles) - 1
	}
	return vizRowStyles[styleIdx].Render(ch)
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
