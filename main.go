package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/ysoftman/findm/internal/player"
	"github.com/ysoftman/findm/internal/tui"
	"github.com/ysoftman/findm/internal/youtube"
)

func main() {
	// Check mpv availability
	if !player.IsAvailable() {
		fmt.Fprintln(os.Stderr, "Warning: mpv is not installed. Playback will not work.")
		fmt.Fprintln(os.Stderr, "Install mpv: brew install mpv")
		fmt.Fprintln(os.Stderr, "")
	}

	// Check yt-dlp availability
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		fmt.Fprintln(os.Stderr, "Warning: yt-dlp is not installed. Search will not work.")
		fmt.Fprintln(os.Stderr, "Install yt-dlp: brew install yt-dlp")
		fmt.Fprintln(os.Stderr, "")
	}

	// Create YouTube client
	client, err := youtube.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "YouTube client error: %v\n", err)
		os.Exit(1)
	}

	// Start TUI
	model := tui.NewModel(client)
	defer model.Cleanup()

	// Handle OS signals to ensure mpv process cleanup
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		<-sigCh
		model.Cleanup()
		os.Exit(0)
	}()

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
