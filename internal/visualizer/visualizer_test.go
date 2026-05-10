package visualizer

import (
	"fmt"
	"math"
	"strings"
	"testing"
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
		"noise_reduction = 82",
	}
	for _, part := range wantParts {
		if !strings.Contains(config, part) {
			t.Fatalf("liveInputConfig() missing %q", part)
		}
	}
}
