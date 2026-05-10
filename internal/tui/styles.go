package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	primaryColor   = lipgloss.Color("#7C3AED")
	secondaryColor = lipgloss.Color("#EC4899")
	mutedColor     = lipgloss.Color("#6B7280")
	successColor   = lipgloss.Color("#10B981")
	warningColor   = lipgloss.Color("#F59E0B")

	// Title style
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor)

	// Tab styles
	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(primaryColor).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Padding(0, 2)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(secondaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#D1D5DB"))

	// Info styles
	channelStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Italic(true)

	durationStyle = lipgloss.NewStyle().
			Foreground(warningColor)

	viewCountStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	urlStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			Underline(true)

	// Player bar style
	playerBarStyle = lipgloss.NewStyle().
			Foreground(successColor).
			Bold(true).
			Border(lipgloss.NormalBorder(), true, false, false, false).
			BorderForeground(mutedColor)

	// Help style
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor).
			MarginTop(1)

	// Status message style
	statusStyle = lipgloss.NewStyle().
			Foreground(successColor)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	// Scroll indicator style
	scrollIndicatorStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Italic(true)

	// Progress bar styles
	progressFilledStyle = lipgloss.NewStyle().
				Foreground(successColor)

	progressEmptyStyle = lipgloss.NewStyle().
				Foreground(mutedColor)

	// Visualizer row palette: cyan tones instead of rainbow bands.
	vizRowStyles = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#155E75")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#0891B2")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#22D3EE")),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#A5F3FC")),
	}
)
