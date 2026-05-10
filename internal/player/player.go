package player

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// State represents the current playback state.
type State int

const (
	Stopped State = iota
	Preparing
	Playing
	Paused
)

const socketPath = "/tmp/findm-mpv.sock"

// Player controls mpv for audio playback via IPC socket.
type Player struct {
	mu         sync.Mutex
	cmd        *exec.Cmd
	state      State
	current    string // current video title
	url        string
	done       chan struct{}
	socketPath string
	conn       net.Conn
	reader     *bufio.Reader
	lastErr    string // last fatal error from mpv (consumed once)
}

// ConsumeError returns and clears the most recent fatal mpv error, if any.
func (p *Player) ConsumeError() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	msg := p.lastErr
	p.lastErr = ""
	return msg
}

// New creates a new Player instance.
func New() *Player {
	return &Player{state: Stopped, socketPath: socketPath}
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

	// Remove stale socket file
	_ = os.Remove(p.socketPath)

	logBuf := &bytes.Buffer{}
	p.cmd = exec.Command("mpv",
		"--no-video",
		"--quiet",
		"--msg-level=all=error",
		"--input-ipc-server="+p.socketPath,
		url,
	)
	// mpv writes its error messages to stdout; capture both streams.
	p.cmd.Stdout = logBuf
	p.cmd.Stderr = logBuf
	if err := p.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start mpv: %w", err)
	}

	p.state = Preparing
	p.current = title
	p.url = url
	p.lastErr = ""
	p.done = make(chan struct{}, 1)
	startedAt := time.Now()

	// Connect IPC socket (non-blocking, retry in background)
	go p.connectIPC()

	// Wait for completion in background
	go func() {
		exitErr := p.cmd.Wait()
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.url == url {
			// If mpv exited before reaching Playing state, surface the error.
			earlyExit := p.state != Playing && p.state != Paused
			if earlyExit && exitErr != nil {
				if msg := extractMpvError(logBuf.String()); msg != "" {
					p.lastErr = msg
				} else if time.Since(startedAt) < 30*time.Second {
					p.lastErr = "mpv exited before playback could start"
				}
			}
			p.closeConn()
			p.state = Stopped
			p.current = ""
			p.url = ""
			p.cmd = nil
		}
		close(p.done)
	}()

	return nil
}

// extractMpvError pulls the most informative line out of mpv's stderr.
func extractMpvError(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Prefer the inner ytdl error line (contains the YouTube reason).
		if strings.Contains(line, "[youtube]") || strings.Contains(line, "ERROR:") {
			line = strings.TrimPrefix(line, "[ytdl_hook] ")
			line = strings.TrimPrefix(line, "ERROR: ")
			return line
		}
	}
	// Fallback: return the first non-empty line.
	for _, line := range strings.Split(s, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// connectIPC connects to the mpv IPC socket with retries.
func (p *Player) connectIPC() {
	// Phase 1: connect to IPC socket
	for range 20 {
		time.Sleep(100 * time.Millisecond)

		p.mu.Lock()
		if p.state == Stopped {
			p.mu.Unlock()
			return
		}
		sp := p.socketPath
		p.mu.Unlock()

		conn, err := net.Dial("unix", sp)
		if err != nil {
			continue
		}

		p.mu.Lock()
		if p.state == Stopped {
			_ = conn.Close()
			p.mu.Unlock()
			return
		}
		p.conn = conn
		p.reader = bufio.NewReader(conn)
		p.mu.Unlock()

		// Phase 2: wait until mpv has loaded the media
		p.waitForMediaReady()
		return
	}
}

// waitForMediaReady waits until mpv reports a valid duration (media loaded).
func (p *Player) waitForMediaReady() {
	for range 50 {
		time.Sleep(200 * time.Millisecond)

		p.mu.Lock()
		if p.state == Stopped {
			p.mu.Unlock()
			return
		}

		data, err := p.getProperty("duration")
		if err != nil {
			p.mu.Unlock()
			continue
		}

		var dur float64
		if err := json.Unmarshal(data, &dur); err != nil || dur <= 0 {
			p.mu.Unlock()
			continue
		}

		p.state = Playing
		p.mu.Unlock()
		return
	}

	// Timeout: transition to Playing anyway so user can try
	p.mu.Lock()
	if p.state == Preparing {
		p.state = Playing
	}
	p.mu.Unlock()
}

// closeConn closes the IPC connection and removes the socket file.
func (p *Player) closeConn() {
	if p.conn != nil {
		_ = p.conn.Close()
		p.conn = nil
		p.reader = nil
	}
	_ = os.Remove(p.socketPath)
}

// Stop stops the current playback.
func (p *Player) Stop() {
	p.mu.Lock()
	if p.cmd == nil || p.cmd.Process == nil {
		p.closeConn()
		p.state = Stopped
		p.current = ""
		p.url = ""
		p.mu.Unlock()
		return
	}

	done := p.done
	p.closeConn()
	_ = p.cmd.Process.Kill()
	p.mu.Unlock()

	// Wait for the background goroutine to finish calling Wait()
	if done != nil {
		<-done
	}
}

// ErrNotReady is returned when playback is still preparing.
var ErrNotReady = fmt.Errorf("playback is preparing")

// IsReady returns true if the IPC connection is established and ready.
func (p *Player) IsReady() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.conn != nil && p.state != Preparing && p.state != Stopped
}

// sendCommand sends a JSON command to mpv via IPC and returns the response data.
func (p *Player) sendCommand(args ...any) (json.RawMessage, error) {
	if p.conn == nil || p.reader == nil {
		return nil, ErrNotReady
	}

	cmd := map[string]any{
		"command": args,
	}
	data, err := json.Marshal(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal command: %w", err)
	}
	data = append(data, '\n')

	if err := p.conn.SetDeadline(time.Now().Add(2 * time.Second)); err != nil {
		return nil, fmt.Errorf("failed to set deadline: %w", err)
	}

	if _, err := p.conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write command: %w", err)
	}

	// Read lines, skipping mpv async event messages (they have "event" field)
	for {
		line, err := p.reader.ReadBytes('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check if this is an event message (skip it)
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(line, &raw); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		if _, isEvent := raw["event"]; isEvent {
			continue // skip async events
		}

		// This is a command response
		var resp struct {
			Data  json.RawMessage `json:"data"`
			Error string          `json:"error"`
		}
		if err := json.Unmarshal(line, &resp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		if resp.Error != "success" {
			return nil, fmt.Errorf("mpv error: %s", resp.Error)
		}

		return resp.Data, nil
	}
}

// getProperty retrieves a property value from mpv.
func (p *Player) getProperty(name string) (json.RawMessage, error) {
	return p.sendCommand("get_property", name)
}

// TogglePause toggles between Playing and Paused states.
func (p *Player) TogglePause() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == Stopped {
		return fmt.Errorf("nothing is playing")
	}
	if p.state == Preparing {
		return ErrNotReady
	}

	_, err := p.sendCommand("cycle", "pause")
	if err != nil {
		return err
	}

	if p.state == Playing {
		p.state = Paused
	} else {
		p.state = Playing
	}
	return nil
}

// Seek seeks by the given number of seconds (positive=forward, negative=backward).
func (p *Player) Seek(seconds float64) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == Stopped {
		return fmt.Errorf("nothing is playing")
	}
	if p.state == Preparing {
		return ErrNotReady
	}

	_, err := p.sendCommand("seek", seconds)
	return err
}

// SetVolume sets the playback volume (0-100).
func (p *Player) SetVolume(vol int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.state == Stopped {
		return fmt.Errorf("nothing is playing")
	}
	if p.state == Preparing {
		return ErrNotReady
	}

	_, err := p.sendCommand("set_property", "volume", vol)
	return err
}

// GetPosition returns the current playback position in seconds.
func (p *Player) GetPosition() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := p.getProperty("time-pos")
	if err != nil {
		return 0
	}

	var pos float64
	if err := json.Unmarshal(data, &pos); err != nil {
		return 0
	}
	return pos
}

// GetDuration returns the total duration in seconds.
func (p *Player) GetDuration() float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := p.getProperty("duration")
	if err != nil {
		return 0
	}

	var dur float64
	if err := json.Unmarshal(data, &dur); err != nil {
		return 0
	}
	return dur
}

// GetVolume returns the current volume level.
func (p *Player) GetVolume() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	data, err := p.getProperty("volume")
	if err != nil {
		return 100
	}

	var vol float64
	if err := json.Unmarshal(data, &vol); err != nil {
		return 100
	}
	return int(vol)
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
	case Preparing:
		return "⏳ Preparing"
	case Playing:
		return "▶ Playing"
	case Paused:
		return "⏸ Paused"
	default:
		return "⏹ Stopped"
	}
}
