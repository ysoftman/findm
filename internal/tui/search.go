package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ysoftman/findm/internal/youtube"
)

// searchMsg is sent when search results are ready.
type searchMsg struct {
	results []youtube.Video
	err     error
}

// recommendMsg is sent when recommendation results are ready.
type recommendMsg struct {
	results []youtube.Video
	err     error
}

func newSearchInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Search music... (e.g. 카페 음악, acoustic chill)"
	ti.CharLimit = 100
	ti.Width = 60
	return ti
}

func performSearch(client *youtube.Client, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Search(query, 10)
		return searchMsg{results: results, err: err}
	}
}

func performRecommend(client *youtube.Client, videoID string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Recommend(videoID, 10)
		return recommendMsg{results: results, err: err}
	}
}
