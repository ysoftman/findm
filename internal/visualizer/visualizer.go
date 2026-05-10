package visualizer

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"math/rand/v2"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	bars             = 48
	rawFrameMaxRange = 1000
	liveInputBinary  = "cava"
	liveInputMethod  = "FINDM_VISUALIZER_METHOD"
	liveInputSource  = "FINDM_VISUALIZER_SOURCE"
)

// Visualizer generates an animated audio visualization effect.
type Visualizer struct {
	mu         sync.Mutex
	values     []float64
	targets    []float64
	peaks      []float64
	running    bool
	stopCh     chan struct{}
	cmd        *exec.Cmd
	configPath string
}

// New creates a new Visualizer instance.
func New() *Visualizer {
	return &Visualizer{
		values:  make([]float64, bars),
		targets: make([]float64, bars),
		peaks:   make([]float64, bars),
	}
}

// Start begins the visualization animation.
func (v *Visualizer) Start() {
	v.mu.Lock()
	if v.running {
		v.mu.Unlock()
		return
	}

	v.running = true
	v.stopCh = make(chan struct{})
	v.resetLocked()
	stopCh := v.stopCh
	v.mu.Unlock()

	if v.startLiveInput(stopCh) {
		return
	}
	go v.animate(stopCh)
}

// Stop stops the visualization.
func (v *Visualizer) Stop() {
	v.mu.Lock()
	if !v.running {
		v.mu.Unlock()
		return
	}

	stopCh := v.stopCh
	cmd := v.cmd
	configPath := v.configPath

	v.running = false
	v.stopCh = nil
	v.cmd = nil
	v.configPath = ""
	if stopCh != nil {
		close(stopCh)
	}
	v.resetLocked()
	v.mu.Unlock()

	if cmd != nil && cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
	if configPath != "" {
		_ = os.Remove(configPath)
	}
}

// IsRunning returns true if the visualizer is active.
func (v *Visualizer) IsRunning() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return v.running
}

// Values returns a copy of the current bar values (0.0 - 1.0).
func (v *Visualizer) Values() []float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]float64, len(v.values))
	copy(out, v.values)
	return out
}

// Peaks returns a copy of the current peak values (0.0 - 1.0).
func (v *Visualizer) Peaks() []float64 {
	v.mu.Lock()
	defer v.mu.Unlock()
	out := make([]float64, len(v.peaks))
	copy(out, v.peaks)
	return out
}

func (v *Visualizer) resetLocked() {
	for i := range v.values {
		v.values[i] = 0
		v.targets[i] = 0
		v.peaks[i] = 0
	}
}

func (v *Visualizer) startLiveInput(stopCh chan struct{}) bool {
	binaryPath, err := exec.LookPath(liveInputBinary)
	if err != nil {
		return false
	}

	configFile, err := os.CreateTemp("", "findm-visualizer-*.conf")
	if err != nil {
		return false
	}
	configPath := configFile.Name()

	if _, err := configFile.WriteString(liveInputConfig()); err != nil {
		_ = configFile.Close()
		_ = os.Remove(configPath)
		return false
	}
	if err := configFile.Close(); err != nil {
		_ = os.Remove(configPath)
		return false
	}

	cmd := exec.Command(binaryPath, "-p", configPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = os.Remove(configPath)
		return false
	}
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		_ = os.Remove(configPath)
		return false
	}

	v.mu.Lock()
	if !v.running || v.stopCh != stopCh {
		v.mu.Unlock()
		_ = cmd.Process.Kill()
		_ = os.Remove(configPath)
		return true
	}
	v.cmd = cmd
	v.configPath = configPath
	v.mu.Unlock()

	go v.readLiveInput(stdout, stopCh)
	go v.waitLiveInput(cmd, configPath, stopCh)
	return true
}

func liveInputConfig() string {
	var input strings.Builder
	method := strings.TrimSpace(os.Getenv(liveInputMethod))
	source := strings.TrimSpace(os.Getenv(liveInputSource))
	if runtime.GOOS == "darwin" {
		if method == "" {
			method = "coreaudio"
		}
		if source == "" {
			source = "auto"
		}
	}
	if method != "" {
		fmt.Fprintf(&input, "method = %s\n", method)
	}
	if source != "" {
		fmt.Fprintf(&input, "source = %s\n", source)
	}

	return fmt.Sprintf(`[general]
framerate = 60
autosens = 1
sensitivity = 110
bars = %d
lower_cutoff_freq = 50
higher_cutoff_freq = 10000
sleep_timer = 0

[input]
%s
[output]
method = raw
channels = mono
mono_option = average
raw_target = /dev/stdout
data_format = ascii
ascii_max_range = %d
bar_delimiter = 59
frame_delimiter = 10

[smoothing]
monstercat = 1
waves = 1
noise_reduction = 82

[eq]
1 = 1.20
2 = 1.05
3 = 0.95
4 = 0.95
5 = 1.05
`, bars, input.String(), rawFrameMaxRange)
}

func (v *Visualizer) readLiveInput(stdout io.Reader, stopCh chan struct{}) {
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 0, 4096), 64*1024)

	for scanner.Scan() {
		select {
		case <-stopCh:
			return
		default:
		}

		frame := parseRawFrame(scanner.Text())
		if len(frame) == 0 {
			continue
		}
		v.applyFrame(frame)
	}
}

func (v *Visualizer) waitLiveInput(cmd *exec.Cmd, configPath string, stopCh chan struct{}) {
	_ = cmd.Wait()
	_ = os.Remove(configPath)

	v.mu.Lock()
	shouldFallback := v.running && v.stopCh == stopCh && v.cmd == cmd
	if shouldFallback {
		v.cmd = nil
		v.configPath = ""
	}
	v.mu.Unlock()

	if shouldFallback {
		go v.animate(stopCh)
	}
}

func parseRawFrame(line string) []float64 {
	parts := strings.Split(strings.TrimSpace(line), ";")
	frame := make([]float64, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		rawValue, err := strconv.Atoi(part)
		if err != nil {
			continue
		}

		value := float64(rawValue) / rawFrameMaxRange
		if value < 0 {
			value = 0
		}
		if value > 1 {
			value = 1
		}
		frame = append(frame, value)
	}
	return frame
}

func (v *Visualizer) applyFrame(frame []float64) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if !v.running {
		return
	}

	for i := range v.values {
		target := resampledFrameValue(frame, i, len(v.values))
		current := v.values[i]

		if target > current {
			v.values[i] = current + (target-current)*0.85
		} else {
			v.values[i] = current + (target-current)*0.45
		}

		if v.values[i] > v.peaks[i] {
			v.peaks[i] = v.values[i]
		} else {
			v.peaks[i] -= 0.045
			if v.peaks[i] < 0 {
				v.peaks[i] = 0
			}
		}
	}
}

func resampledFrameValue(frame []float64, idx, total int) float64 {
	if len(frame) == 0 || total <= 0 {
		return 0
	}
	if len(frame) == 1 || total == 1 {
		return frame[0]
	}

	position := float64(idx) * float64(len(frame)-1) / float64(total-1)
	left := int(math.Floor(position))
	right := left + 1
	if right >= len(frame) {
		return frame[left]
	}

	blend := position - float64(left)
	return frame[left]*(1-blend) + frame[right]*blend
}

func (v *Visualizer) animate(stopCh chan struct{}) {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()

	beatTick := time.NewTicker(170 * time.Millisecond)
	defer beatTick.Stop()

	accentTick := time.NewTicker(800 * time.Millisecond)
	defer accentTick.Stop()

	frame := 0
	energy := 0.7

	for {
		select {
		case <-stopCh:
			return

		case <-accentTick.C:
			energy = 0.5 + rand.Float64()*0.5

			v.mu.Lock()
			center := rand.IntN(bars)
			spread := 2 + rand.IntN(4)
			for i := center - spread; i <= center+spread; i++ {
				if i >= 0 && i < bars {
					v.targets[i] = 0.8 + rand.Float64()*0.2
				}
			}
			v.mu.Unlock()

		case <-beatTick.C:
			nextTargets := make([]float64, bars)
			v.mu.Lock()
			for i := range nextTargets {
				fi := float64(i)
				ff := float64(frame)

				base := 0.0
				base += 0.3 * (1 + math.Sin(fi*0.4+ff*0.3))
				base += 0.2 * (1 + math.Sin(fi*1.1-ff*0.15))
				base += 0.15 * (1 + math.Cos(fi*0.7+ff*0.22))

				spike := rand.Float64()
				if spike > 0.6 {
					base += (spike - 0.6) * 2.5
				}
				if i < bars/4 {
					base *= 1.2
				}

				base *= energy
				if base > 1.0 {
					base = 1.0
				}
				if base < 0.02 {
					base = 0.02
				}
				nextTargets[i] = base
			}

			for i := range v.targets {
				smoothed := nextTargets[i] * 0.58
				if i > 0 {
					smoothed += nextTargets[i-1] * 0.21
				} else {
					smoothed += nextTargets[i] * 0.21
				}
				if i < len(nextTargets)-1 {
					smoothed += nextTargets[i+1] * 0.21
				} else {
					smoothed += nextTargets[i] * 0.21
				}
				v.targets[i] = smoothed
			}
			v.mu.Unlock()

		case <-ticker.C:
			v.mu.Lock()
			for i := range v.values {
				target := v.targets[i]
				current := v.values[i]

				if target > current {
					v.values[i] = current + (target-current)*0.45
				} else {
					v.values[i] = current + (target-current)*0.16
				}

				v.values[i] += (rand.Float64() - 0.5) * 0.025
				if v.values[i] > 1.0 {
					v.values[i] = 1.0
				}
				if v.values[i] < 0.0 {
					v.values[i] = 0.0
				}

				if v.values[i] > v.peaks[i] {
					v.peaks[i] = v.values[i]
				} else {
					v.peaks[i] -= 0.03
					if v.peaks[i] < 0 {
						v.peaks[i] = 0
					}
				}
			}
			v.mu.Unlock()
			frame++
		}
	}
}
