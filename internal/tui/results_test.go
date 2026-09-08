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

func TestCurrentVideoID(t *testing.T) {
	m := Model{view: ResultsView, cursor: 1, results: []youtube.Video{
		{ID: "vid1", Kind: youtube.KindVideo},
		{ID: "pl1", Kind: youtube.KindPlaylist},
	}}
	if got := m.currentVideoID(); got != "" {
		t.Fatalf("currentVideoID() on playlist = %q, want empty", got)
	}
	m.cursor = 0
	if got := m.currentVideoID(); got != "vid1" {
		t.Fatalf("currentVideoID() = %q, want vid1", got)
	}
}
