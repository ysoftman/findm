package visualizer

import (
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

const bars = 32

// Visualizer generates an animated audio visualization effect.
type Visualizer struct {
	mu      sync.Mutex
	values  []float64
	targets []float64
	peaks   []float64
	running bool
	stopCh  chan struct{}
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
	defer v.mu.Unlock()

	if v.running {
		return
	}

	v.running = true
	v.stopCh = make(chan struct{})

	go v.animate()
}

// Stop stops the visualization.
func (v *Visualizer) Stop() {
	v.mu.Lock()
	if !v.running {
		v.mu.Unlock()
		return
	}
	v.running = false
	close(v.stopCh)
	for i := range v.values {
		v.values[i] = 0
		v.targets[i] = 0
		v.peaks[i] = 0
	}
	v.mu.Unlock()
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

func (v *Visualizer) animate() {
	// Fast tick for smooth interpolation
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	// Beat generator on separate rhythm
	beatTick := time.NewTicker(400 * time.Millisecond)
	defer beatTick.Stop()

	// Occasional accent beats
	accentTick := time.NewTicker(1600 * time.Millisecond)
	defer accentTick.Stop()

	frame := 0
	energy := 0.7 // overall energy level

	for {
		select {
		case <-v.stopCh:
			return

		case <-accentTick.C:
			// Shift overall energy for variety
			energy = 0.5 + rand.Float64()*0.5

			// Accent: spike a random cluster of bars
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
			// Generate new targets simulating a beat
			v.mu.Lock()
			for i := range v.targets {
				fi := float64(i)
				ff := float64(frame)

				// Base: multiple waves at different frequencies
				base := 0.0
				base += 0.3 * (1 + math.Sin(fi*0.4+ff*0.3))
				base += 0.2 * (1 + math.Sin(fi*1.1-ff*0.15))
				base += 0.15 * (1 + math.Cos(fi*0.7+ff*0.22))

				// Random spike per bar (simulates different frequency bands)
				spike := rand.Float64()
				if spike > 0.6 {
					base += (spike - 0.6) * 2.5
				}

				// Low frequencies (left bars) tend to be stronger
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
				v.targets[i] = base
			}
			v.mu.Unlock()

		case <-ticker.C:
			v.mu.Lock()
			for i := range v.values {
				target := v.targets[i]
				current := v.values[i]

				// Fast attack, slower decay (like real audio meters)
				if target > current {
					v.values[i] = current + (target-current)*0.6
				} else {
					v.values[i] = current + (target-current)*0.2
				}

				// Add micro-jitter for liveliness
				v.values[i] += (rand.Float64() - 0.5) * 0.08

				// Clamp
				if v.values[i] > 1.0 {
					v.values[i] = 1.0
				}
				if v.values[i] < 0.0 {
					v.values[i] = 0.0
				}

				// Peak hold with slow decay
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
