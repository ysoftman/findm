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

	// Visualizer gradient: cyan → green → yellow → orange → hot pink (16 steps, vibrant neon palette)
	vizGradient = []lipgloss.Style{
		lipgloss.NewStyle().Foreground(lipgloss.Color("#06D6A0")), // mint
		lipgloss.NewStyle().Foreground(lipgloss.Color("#1BD47E")), // mint-green
		lipgloss.NewStyle().Foreground(lipgloss.Color("#36D25C")), // green
		lipgloss.NewStyle().Foreground(lipgloss.Color("#56CF3A")), // lime-green
		lipgloss.NewStyle().Foreground(lipgloss.Color("#7ACC18")), // lime
		lipgloss.NewStyle().Foreground(lipgloss.Color("#A0C800")), // yellow-green
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C4BE00")), // gold
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E0AE00")), // amber
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")), // orange-yellow
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F97316")), // orange
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FB5C3C")), // coral
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F94468")), // rose
		lipgloss.NewStyle().Foreground(lipgloss.Color("#F43090")), // pink
		lipgloss.NewStyle().Foreground(lipgloss.Color("#E826B0")), // magenta
		lipgloss.NewStyle().Foreground(lipgloss.Color("#D624D0")), // purple-pink
		lipgloss.NewStyle().Foreground(lipgloss.Color("#C026E8")), // violet
	}

	vizPeakStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#C026E8")) // same as highest level
)
