package visualizer

import (
	"fmt"
	"math"
	"strings"
	"testing"
	"time"
)

func TestParseRawFrame(t *testing.T) {
	got := parseRawFrame("0;500;1000;1200;-1;bad;\n")
	want := []float64{0, 0.5, 1, 1, 0}

	if len(got) != len(want) {
		t.Fatalf("len(parseRawFrame()) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(got[i]-want[i]) > 0.0001 {
			t.Fatalf("parseRawFrame()[%d] = %f, want %f", i, got[i], want[i])
		}
	}
}

func TestResampledFrameValue(t *testing.T) {
	got := resampledFrameValue([]float64{0, 1}, 1, 3)
	if math.Abs(got-0.5) > 0.0001 {
		t.Fatalf("resampledFrameValue() = %f, want 0.5", got)
	}
}

func TestLiveInputConfig(t *testing.T) {
	t.Setenv(liveInputMethod, "fifo")
	t.Setenv(liveInputSource, "/tmp/findm-test.fifo")

	config := liveInputConfig()
	wantParts := []string{
		fmt.Sprintf("bars = %d", bars),
		"method = fifo",
		"source = /tmp/findm-test.fifo",
		"method = raw",
		"data_format = ascii",
		fmt.Sprintf("ascii_max_range = %d", rawFrameMaxRange),
		"noise_reduction = 77",
	}
	for _, part := range wantParts {
		if !strings.Contains(config, part) {
			t.Fatalf("liveInputConfig() missing %q", part)
		}
	}
}

func TestStepLockedEasesTowardTargets(t *testing.T) {
	v := New()
	v.targets[0] = 1
	v.values[1] = 1 // target 0: falls

	v.stepLocked(16 * time.Millisecond)
	rise, fall := v.values[0], v.values[1]
	if rise <= 0 || rise >= 1 {
		t.Fatalf("rising bar = %f, want strictly between 0 and 1", rise)
	}
	if fall <= 0 || fall >= 1 {
		t.Fatalf("falling bar = %f, want strictly between 0 and 1", fall)
	}
	if 1-fall >= rise {
		t.Fatalf("fall moved %f but rise moved %f; attack should be faster than decay", 1-fall, rise)
	}

	for range 200 {
		v.stepLocked(16 * time.Millisecond)
	}
	if math.Abs(v.values[0]-1) > 0.001 || v.values[1] > 0.001 {
		t.Fatalf("values after settling = %v, want [1 0 ...]", v.values[:2])
	}
}
