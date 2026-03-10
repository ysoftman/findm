package tui

import (
	"fmt"

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
	maxTitle := width - 30
	if maxTitle < 20 {
		maxTitle = 20
	}
	if len(title) > maxTitle {
		title = title[:maxTitle-3] + "..."
	}

	status := p.StateString()
	return style.Render(fmt.Sprintf("  %s  %s  |  Space: stop", status, title))
}
