package player

import (
	"fmt"
	"os/exec"
	"sync"
)

// State represents the current playback state.
type State int

const (
	Stopped State = iota
	Playing
)

// Player controls mpv for audio playback.
type Player struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	state   State
	current string // current video title
	url     string
	done    chan struct{}
}

// New creates a new Player instance.
func New() *Player {
	return &Player{state: Stopped}
}

// IsAvailable checks if mpv is installed.
func IsAvailable() bool {
	_, err := exec.LookPath("mpv")
	return err == nil
}

// Play starts playing the given YouTube URL.
func (p *Player) Play(url, title string) error {
	if !IsAvailable() {
		return fmt.Errorf("mpv is not installed. Install it with: brew install mpv")
	}

	p.Stop()

	p.mu.Lock()
	defer p.mu.Unlock()

	p.cmd = exec.Command("mpv", "--no-video", "--really-quiet", url)
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	p.state = Playing
	p.current = title
	p.url = url
	p.done = make(chan struct{}, 1)

	// Wait for completion in background — only this goroutine calls Wait()
	go func() {
		_ = p.cmd.Wait()
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.url == url {
			p.state = Stopped
			p.current = ""
			p.url = ""
			p.cmd = nil
		}
		close(p.done)
	}()

	return nil
}

// Stop stops the current playback.
func (p *Player) Stop() {
	p.mu.Lock()
	if p.cmd == nil || p.cmd.Process == nil {
		p.state = Stopped
		p.current = ""
		p.url = ""
		p.mu.Unlock()
		return
	}

	done := p.done
	_ = p.cmd.Process.Kill()
	p.mu.Unlock()

	// Wait for the background goroutine to finish calling Wait()
	if done != nil {
		<-done
	}
}

// GetState returns the current playback state.
func (p *Player) GetState() State {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.state
}

// CurrentTitle returns the title of the currently playing track.
func (p *Player) CurrentTitle() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.current
}

// StateString returns a human-readable state string.
func (p *Player) StateString() string {
	switch p.GetState() {
	case Playing:
		return "▶ Playing"
	default:
		return "⏹ Stopped"
	}
}
