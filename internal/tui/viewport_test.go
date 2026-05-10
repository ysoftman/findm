package tui

import "testing"

func TestViewportBoundsKeepsCursorVisible(t *testing.T) {
	start, end := viewportBounds(10, 7, 4)
	if start != 4 || end != 8 {
		t.Fatalf("viewportBounds() = (%d, %d), want (4, 8)", start, end)
	}
}

func TestFittedVisibleCountIncludesIndicatorLines(t *testing.T) {
	got := fittedVisibleCount(10, 5, 5, func(start, end int) int {
		lines := end - start
		if start > 0 {
			lines++
		}
		if end < 10 {
			lines++
		}
		return lines
	})
	if got != 3 {
		t.Fatalf("fittedVisibleCount() = %d, want 3", got)
	}
}

func TestRenderedLineCount(t *testing.T) {
	for input, want := range map[string]int{
		"":        0,
		"one":     1,
		"one\n":   1,
		"one\n\n": 2,
	} {
		if got := renderedLineCount(input); got != want {
			t.Fatalf("renderedLineCount(%q) = %d, want %d", input, got, want)
		}
	}
}
