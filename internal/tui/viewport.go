package tui

import "strings"

func renderedLineCount(s string) int {
	if s == "" {
		return 0
	}

	lines := strings.Count(s, "\n")
	if !strings.HasSuffix(s, "\n") {
		lines++
	}
	return lines
}

func viewportBounds(total, cursor, visibleCount int) (int, int) {
	if total <= 0 || visibleCount <= 0 {
		return 0, 0
	}
	if visibleCount > total {
		visibleCount = total
	}
	if cursor < 0 {
		cursor = 0
	}
	if cursor >= total {
		cursor = total - 1
	}

	start := 0
	if cursor >= visibleCount {
		start = cursor - visibleCount + 1
	}
	if start+visibleCount > total {
		start = total - visibleCount
	}
	if start < 0 {
		start = 0
	}

	return start, start + visibleCount
}

func fittedVisibleCount(total, cursor, height int, lineCount func(start, end int) int) int {
	if total <= 0 {
		return 0
	}
	if height <= 0 {
		return total
	}

	for visibleCount := total; visibleCount >= 1; visibleCount-- {
		start, end := viewportBounds(total, cursor, visibleCount)
		if lineCount(start, end) <= height {
			return visibleCount
		}
	}

	return 1
}
