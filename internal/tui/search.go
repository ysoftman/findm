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

// searchMoreMsg is sent when more search results are ready.
type searchMoreMsg struct {
	results []youtube.Video
	err     error
}

// recommendMsg is sent when recommendation results are ready.
type recommendMsg struct {
	results []youtube.Video
	err     error
}

// expandMsg is sent when a playlist/channel's videos have been loaded.
type expandMsg struct {
	source  string
	results []youtube.Video
	err     error
}

func newSearchInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = "Search music, @handle, publisher, or URL..."
	ti.CharLimit = 100
	ti.Width = 60
	return ti
}

func performSearch(client *youtube.Client, query string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Search(query, 10, 0)
		return searchMsg{results: results, err: err}
	}
}

func performSearchMore(client *youtube.Client, query string, offset int64) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Search(query, 10, offset)
		return searchMoreMsg{results: results, err: err}
	}
}

func performRecommend(client *youtube.Client, videoID string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.Recommend(videoID, 10)
		return recommendMsg{results: results, err: err}
	}
}

func performExpand(client *youtube.Client, v youtube.Video) tea.Cmd {
	return performExpandRaw(client, v.URL, v.Title)
}

func performExpandRaw(client *youtube.Client, url, source string) tea.Cmd {
	return func() tea.Msg {
		results, err := client.FetchVideosFromURL(url, 100)
		return expandMsg{source: source, results: results, err: err}
	}
}
