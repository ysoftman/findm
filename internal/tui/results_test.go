package tui

import (
	"strings"
	"testing"

	"github.com/ysoftman/findm/internal/youtube"
)

func TestRenderResultsShowsLoadMoreItem(t *testing.T) {
	output := renderResults([]youtube.Video{
		{Title: "First", Channel: "Channel", Duration: "1:00"},
	}, 1, 10, true)

	if !strings.Contains(output, "Load more results...") {
		t.Fatalf("renderResults() missing load more item:\n%s", output)
	}
}

func TestRenderResultsHidesLoadMoreItem(t *testing.T) {
	output := renderResults([]youtube.Video{
		{Title: "First", Channel: "Channel", Duration: "1:00"},
	}, 0, 10, false)

	if strings.Contains(output, "Load more results...") {
		t.Fatalf("renderResults() unexpectedly included load more item:\n%s", output)
	}
}
