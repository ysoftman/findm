package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/ysoftman/findm/internal/player"
	"github.com/ysoftman/findm/internal/playlist"
	"github.com/ysoftman/findm/internal/visualizer"
	"github.com/ysoftman/findm/internal/youtube"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func vizTickCmd() tea.Cmd {
	return tea.Tick(80*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// handlePlayerErr handles player errors, showing ErrNotReady as status instead of error.
func (m *Model) handlePlayerErr(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, player.ErrNotReady) {
		m.statusMsg = preparingPlaybackMsg
	} else {
		m.errorMsg = err.Error()
	}
}

// playTrackAt plays a track from the current playlist at the given index.
func (m *Model) playTrackAt(idx int) tea.Cmd {
	if m.currentPlaylist == nil || idx < 0 || idx >= len(m.currentPlaylist.Tracks) {
		return nil
	}
	t := m.currentPlaylist.Tracks[idx]
	if err := m.player.Play(t.URL, t.Title); err != nil {
		m.errorMsg = err.Error()
		return nil
	}
	m.playingTrackIdx = idx
	m.trackCursor = idx
	m.autoPlay = true
	m.statusMsg = ""
	return tickCmd()
}

// View represents the current active view.
type View int

const (
	SearchView View = iota
	ResultsView
	PlaylistListView
	PlaylistDetailView
)

// Tab represents the top-level tab.
type Tab int

const (
	SearchTab Tab = iota
	PlaylistTab
)

// Model is the main bubbletea model.
type Model struct {
	client  *youtube.Client
	player  *player.Player
	viz     *visualizer.Visualizer
	version string

	// UI state
	tab         Tab
	view        View
	searchInput textinput.Model
	width       int
	height      int
	animFrame   int

	// Search results
	results        []youtube.Video
	searchQuery    string
	cursor         int
	hasMoreResults bool

	// Playlist state
	playlists       []string
	playlistCursor  int
	currentPlaylist *playlist.Playlist
	trackCursor     int
	playingTrackIdx int  // index of currently playing track in playlist (-1 = none)
	autoPlay        bool // auto-play next track when current finishes

	// Status
	statusMsg string
	errorMsg  string
	loading   bool

	// Playlist creation
	creatingPlaylist bool
	newPlaylistInput textinput.Model

	// Adding to playlist
	addingToPlaylist  bool
	addPlaylistCursor int
}

// NewModel creates a new TUI model.
func NewModel(client *youtube.Client, version string) Model {
	si := newSearchInput()
	si.Focus()

	pi := textinput.New()
	pi.Placeholder = "Playlist name..."
	pi.CharLimit = 50
	pi.Width = 40

	return Model{
		client:           client,
		player:           player.New(),
		viz:              visualizer.New(),
		version:          version,
		playingTrackIdx:  -1,
		tab:              SearchTab,
		view:             SearchView,
		searchInput:      si,
		newPlaylistInput: pi,
	}
}

// Cleanup releases resources (stops mpv process and visualizer).
func (m Model) Cleanup() {
	m.player.Stop()
	m.viz.Stop()
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case searchMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.results = msg.results
		m.cursor = 0
		m.view = ResultsView
		m.hasMoreResults = len(msg.results) > 0
		m.statusMsg = fmt.Sprintf("Found %d results", len(msg.results))
		return m, nil

	case searchMoreMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		if len(msg.results) > 0 {
			m.results = append(m.results, msg.results...)
			m.hasMoreResults = true
			m.statusMsg = fmt.Sprintf("Loaded %d more results (total: %d)", len(msg.results), len(m.results))
		} else {
			m.hasMoreResults = false
			if m.cursor >= len(m.results) && len(m.results) > 0 {
				m.cursor = len(m.results) - 1
			}
			m.statusMsg = "No more results available"
		}
		return m, nil

	case recommendMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		if len(msg.results) == 0 {
			m.statusMsg = "Found 0 recommendations"
			return m, nil
		}
		m.results = msg.results
		m.cursor = 0
		m.view = ResultsView
		m.searchQuery = ""
		m.hasMoreResults = false
		m.statusMsg = fmt.Sprintf("Found %d recommendations", len(msg.results))
		return m, nil

	case expandMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		if len(msg.results) == 0 {
			m.statusMsg = "No videos found"
			return m, nil
		}
		m.results = msg.results
		m.cursor = 0
		m.view = ResultsView
		m.searchQuery = ""
		m.hasMoreResults = false
		m.statusMsg = fmt.Sprintf("Showing %d videos from %s", len(msg.results), msg.source)
		return m, nil

	case tickMsg:
		m.animFrame++
		state := m.player.GetState()
		if state != player.Stopped {
			switch state {
			case player.Preparing:
				m.statusMsg = preparingPlaybackMsg
				m.viz.Stop()
			case player.Playing:
				if m.statusMsg == preparingPlaybackMsg {
					m.statusMsg = ""
				}
				if !m.viz.IsRunning() {
					m.viz.Start()
				}
			}
			if state == player.Preparing || m.viz.IsRunning() {
				return m, vizTickCmd()
			}
			return m, tickCmd()
		}
		// Player stopped — surface any latched error from mpv.
		if errMsg := m.player.ConsumeError(); errMsg != "" {
			m.errorMsg = errMsg
			m.statusMsg = ""
		}
		// Track ended: auto-play next in playlist
		if m.autoPlay && m.currentPlaylist != nil {
			next := m.playingTrackIdx + 1
			if next < len(m.currentPlaylist.Tracks) {
				cmd := m.playTrackAt(next)
				if cmd != nil {
					return m, cmd
				}
			} else {
				m.autoPlay = false
				m.playingTrackIdx = -1
				m.statusMsg = "Playlist finished"
			}
		}
		m.viz.Stop()
		return m, nil

	case playlistsLoadedMsg:
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.playlists = msg.names
		if m.playlistCursor >= len(m.playlists) && m.playlistCursor > 0 {
			m.playlistCursor = len(m.playlists) - 1
		}
		if m.addPlaylistCursor >= len(m.playlists) && m.addPlaylistCursor > 0 {
			m.addPlaylistCursor = len(m.playlists) - 1
		}
		return m, nil

	case tea.KeyMsg:
		// Clear error on any key
		m.errorMsg = ""

		// Handle adding to playlist (check before creatingPlaylist
		// because addToPlaylist has its own creatingPlaylist sub-state)
		if m.addingToPlaylist {
			return m.handleAddToPlaylist(msg)
		}

		// Handle playlist creation input
		if m.creatingPlaylist {
			return m.handlePlaylistCreation(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			if m.view == SearchView && m.searchInput.Focused() {
				if msg.String() == "q" {
					break // let 'q' pass through to text input
				}
			}
			m.player.Stop()
			m.viz.Stop()
			return m, tea.Quit

		case "tab":
			if m.tab == SearchTab {
				m.tab = PlaylistTab
				m.view = PlaylistListView
				m.searchInput.Blur()
				return m, m.loadPlaylists()
			}
			m.tab = SearchTab
			if len(m.results) > 0 {
				m.view = ResultsView
			} else {
				m.view = SearchView
				m.searchInput.Focus()
			}
			return m, nil

		case "/":
			if m.view != SearchView || !m.searchInput.Focused() {
				m.tab = SearchTab
				m.view = SearchView
				m.searchInput.Focus()
				m.searchInput.SetValue("")
				return m, nil
			}
		}

		// Route to view-specific handlers
		switch m.view {
		case SearchView:
			return m.handleSearchView(msg)
		case ResultsView:
			return m.handleResultsView(msg)
		case PlaylistListView:
			return m.handlePlaylistListView(msg)
		case PlaylistDetailView:
			return m.handlePlaylistDetailView(msg)
		}
	}

	// Update text inputs for non-key messages (e.g. Blink)
	if m.creatingPlaylist {
		var cmd tea.Cmd
		m.newPlaylistInput, cmd = m.newPlaylistInput.Update(msg)
		return m, cmd
	}

	if m.view == SearchView && m.searchInput.Focused() {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleSearchView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil
		}
		m.searchInput.Blur()

		if youtube.IsYouTubeWatchURL(query) {
			if err := m.player.Play(query, query); err != nil {
				m.errorMsg = err.Error()
				return m, nil
			}
			m.autoPlay = false
			m.playingTrackIdx = -1
			m.statusMsg = ""
			m.view = ResultsView
			return m, tickCmd()
		}
		if youtube.IsExpandableURL(query) {
			m.loading = true
			m.statusMsg = "Loading..."
			return m, performExpandRaw(m.client, query, query)
		}
		if youtube.IsChannelHandle(query) {
			m.loading = true
			m.statusMsg = "Loading channel..."
			return m, performExpandRaw(m.client, youtube.ChannelHandleURL(query), query)
		}

		m.loading = true
		m.statusMsg = "Searching..."
		m.searchQuery = query
		return m, performSearch(m.client, query)
	case "esc":
		if len(m.results) > 0 {
			m.view = ResultsView
			m.searchInput.Blur()
		}
		return m, nil
	}

	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) canLoadMoreResults() bool {
	return m.searchQuery != "" && m.hasMoreResults && len(m.results) > 0 && !m.loading
}

func (m Model) isLoadMoreSelected() bool {
	return m.canLoadMoreResults() && m.cursor == len(m.results)
}

func (m Model) resultCursorMax() int {
	if m.canLoadMoreResults() {
		return len(m.results)
	}
	if len(m.results) == 0 {
		return 0
	}
	return len(m.results) - 1
}

func (m *Model) loadMoreResults() tea.Cmd {
	if !m.canLoadMoreResults() {
		return nil
	}
	m.loading = true
	m.statusMsg = "Loading more results..."
	offset := int64(len(m.results))
	return performSearchMore(m.client, m.searchQuery, offset)
}

func (m Model) handleResultsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < m.resultCursorMax() {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if m.isLoadMoreSelected() {
			return m, m.loadMoreResults()
		}
		if len(m.results) > 0 && m.cursor < len(m.results) {
			v := m.results[m.cursor]
			if !v.Kind.Playable() {
				m.loading = true
				m.statusMsg = fmt.Sprintf("Loading %s...", kindLabel(v.Kind))
				return m, performExpand(m.client, v)
			}
			if err := m.player.Play(v.URL, v.Title); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.autoPlay = false
				m.playingTrackIdx = -1
				m.statusMsg = ""
				return m, tickCmd()
			}
		}
	case "n":
		if next := nextPlayableIdx(m.results, m.cursor, +1); next >= 0 {
			m.cursor = next
			v := m.results[m.cursor]
			if err := m.player.Play(v.URL, v.Title); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.autoPlay = false
				m.playingTrackIdx = -1
				m.statusMsg = ""
				return m, tickCmd()
			}
		}
	case "p":
		if prev := nextPlayableIdx(m.results, m.cursor, -1); prev >= 0 {
			m.cursor = prev
			v := m.results[m.cursor]
			if err := m.player.Play(v.URL, v.Title); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.autoPlay = false
				m.playingTrackIdx = -1
				m.statusMsg = ""
				return m, tickCmd()
			}
		}
	case " ":
		m.handlePlayerErr(m.player.TogglePause())
		if m.player.GetState() == player.Paused {
			m.viz.Stop()
		} else if m.player.GetState() == player.Playing {
			m.viz.Start()
		}
	case "s":
		m.player.Stop()
		m.viz.Stop()
		m.autoPlay = false
		m.playingTrackIdx = -1
		m.statusMsg = "Playback stopped"
	case "left", "h":
		m.handlePlayerErr(m.player.Seek(-10))
	case "right", "l":
		m.handlePlayerErr(m.player.Seek(10))
	case "+", "=":
		vol := m.player.GetVolume() + 10
		m.handlePlayerErr(m.player.SetVolume(vol))
	case "-":
		vol := m.player.GetVolume() - 10
		m.handlePlayerErr(m.player.SetVolume(vol))
	case "r":
		if len(m.results) > 0 && m.cursor < len(m.results) {
			v := m.results[m.cursor]
			if !v.Kind.Playable() {
				m.statusMsg = "Recommend works on videos only"
				return m, nil
			}
			m.loading = true
			m.statusMsg = "Finding recommendations..."
			return m, performRecommend(m.client, v.ID)
		}
	case "L":
		if cmd := m.loadMoreResults(); cmd != nil {
			return m, cmd
		}
	case "a":
		if len(m.results) > 0 && m.cursor < len(m.results) {
			v := m.results[m.cursor]
			if !v.Kind.Playable() {
				m.statusMsg = "Only videos can be added to playlists"
				return m, nil
			}
			m.addingToPlaylist = true
			m.addPlaylistCursor = 0
			return m, m.loadPlaylists()
		}
	}
	return m, nil
}

func (m Model) handlePlaylistListView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.playlistCursor < len(m.playlists)-1 {
			m.playlistCursor++
		}
	case "k", "up":
		if m.playlistCursor > 0 {
			m.playlistCursor--
		}
	case "enter":
		if len(m.playlists) > 0 {
			name := m.playlists[m.playlistCursor]
			pl, err := playlist.Get(name)
			if err != nil {
				m.errorMsg = err.Error()
				return m, nil
			}
			m.currentPlaylist = pl
			m.trackCursor = 0
			m.view = PlaylistDetailView
		}
	case "c":
		m.creatingPlaylist = true
		m.newPlaylistInput.SetValue("")
		m.newPlaylistInput.Focus()
		return m, nil
	case "d":
		if len(m.playlists) > 0 {
			name := m.playlists[m.playlistCursor]
			if err := playlist.Delete(name); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Deleted playlist: %s", name)
				if m.playlistCursor > 0 {
					m.playlistCursor--
				}
			}
			return m, m.loadPlaylists()
		}
	case "esc":
		m.tab = SearchTab
		if len(m.results) > 0 {
			m.view = ResultsView
		} else {
			m.view = SearchView
			m.searchInput.Focus()
		}
	}
	return m, nil
}

func (m Model) handlePlaylistDetailView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.currentPlaylist != nil && m.trackCursor < len(m.currentPlaylist.Tracks)-1 {
			m.trackCursor++
		}
	case "k", "up":
		if m.trackCursor > 0 {
			m.trackCursor--
		}
	case "enter":
		if m.currentPlaylist != nil && len(m.currentPlaylist.Tracks) > 0 {
			cmd := m.playTrackAt(m.trackCursor)
			if cmd != nil {
				return m, cmd
			}
		}
	case "n":
		if m.currentPlaylist != nil && m.playingTrackIdx >= 0 {
			next := m.playingTrackIdx + 1
			if next < len(m.currentPlaylist.Tracks) {
				cmd := m.playTrackAt(next)
				if cmd != nil {
					return m, cmd
				}
			}
		}
	case "p":
		if m.currentPlaylist != nil && m.playingTrackIdx > 0 {
			cmd := m.playTrackAt(m.playingTrackIdx - 1)
			if cmd != nil {
				return m, cmd
			}
		}
	case " ":
		m.handlePlayerErr(m.player.TogglePause())
		if m.player.GetState() == player.Paused {
			m.viz.Stop()
		} else if m.player.GetState() == player.Playing {
			m.viz.Start()
		}
	case "s":
		m.player.Stop()
		m.viz.Stop()
		m.autoPlay = false
		m.playingTrackIdx = -1
		m.statusMsg = "Playback stopped"
	case "left":
		m.handlePlayerErr(m.player.Seek(-10))
	case "right":
		m.handlePlayerErr(m.player.Seek(10))
	case "+", "=":
		vol := m.player.GetVolume() + 10
		m.handlePlayerErr(m.player.SetVolume(vol))
	case "-":
		vol := m.player.GetVolume() - 10
		m.handlePlayerErr(m.player.SetVolume(vol))
	case "d":
		if m.currentPlaylist != nil && len(m.currentPlaylist.Tracks) > 0 {
			if err := playlist.RemoveTrack(m.currentPlaylist.Name, m.trackCursor); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = "Track removed"
				pl, err := playlist.Get(m.currentPlaylist.Name)
				if err != nil {
					m.errorMsg = err.Error()
					return m, nil
				}
				m.currentPlaylist = pl
				if m.trackCursor >= len(m.currentPlaylist.Tracks) && m.trackCursor > 0 {
					m.trackCursor--
				}
			}
		}
	case "esc":
		m.view = PlaylistListView
		return m, m.loadPlaylists()
	}
	return m, nil
}

func (m Model) handlePlaylistCreation(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.newPlaylistInput.Value())
		if name != "" {
			if err := playlist.Create(name); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Created playlist: %s", name)
			}
		}
		m.creatingPlaylist = false
		m.newPlaylistInput.Blur()
		return m, m.loadPlaylists()
	case "esc":
		m.creatingPlaylist = false
		m.newPlaylistInput.Blur()
		return m, nil
	default:
		var cmd tea.Cmd
		m.newPlaylistInput, cmd = m.newPlaylistInput.Update(msg)
		return m, cmd
	}
}

func (m Model) handleAddToPlaylist(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If creating a new playlist within add overlay
	if m.creatingPlaylist {
		switch msg.String() {
		case "enter":
			name := strings.TrimSpace(m.newPlaylistInput.Value())
			if name != "" {
				if err := playlist.Create(name); err != nil {
					m.errorMsg = err.Error()
				} else {
					// Created and auto-add the track
					if len(m.results) > 0 {
						v := m.results[m.cursor]
						track := playlist.Track{
							VideoID: v.ID,
							Title:   v.Title,
							Channel: v.Channel,
							URL:     v.URL,
						}
						if err := playlist.AddTrack(name, track); err != nil {
							m.errorMsg = err.Error()
						} else {
							m.statusMsg = fmt.Sprintf("Created %s and added: %s", name, v.Title)
						}
					}
				}
			}
			m.creatingPlaylist = false
			m.addingToPlaylist = false
			m.newPlaylistInput.Blur()
			return m, m.loadPlaylists()
		case "esc":
			m.creatingPlaylist = false
			m.newPlaylistInput.Blur()
			return m, nil
		default:
			var cmd tea.Cmd
			m.newPlaylistInput, cmd = m.newPlaylistInput.Update(msg)
			return m, cmd
		}
	}

	switch msg.String() {
	case "j", "down":
		if m.addPlaylistCursor < len(m.playlists)-1 {
			m.addPlaylistCursor++
		}
	case "k", "up":
		if m.addPlaylistCursor > 0 {
			m.addPlaylistCursor--
		}
	case "enter":
		if len(m.playlists) > 0 && len(m.results) > 0 {
			v := m.results[m.cursor]
			track := playlist.Track{
				VideoID: v.ID,
				Title:   v.Title,
				Channel: v.Channel,
				URL:     v.URL,
			}
			name := m.playlists[m.addPlaylistCursor]
			if err := playlist.AddTrack(name, track); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = fmt.Sprintf("Added to %s: %s", name, v.Title)
			}
		}
		m.addingToPlaylist = false
	case "c":
		m.creatingPlaylist = true
		m.newPlaylistInput.SetValue("")
		m.newPlaylistInput.Focus()
		return m, nil
	case "esc":
		m.addingToPlaylist = false
	}
	return m, nil
}

type playlistsLoadedMsg struct {
	names []string
	err   error
}

func (m Model) loadPlaylists() tea.Cmd {
	return func() tea.Msg {
		names, err := playlist.List()
		return playlistsLoadedMsg{names: names, err: err}
	}
}

func (m Model) View() string {
	var sb strings.Builder

	// Title and Tabs (same line)
	searchTab := inactiveTabStyle.Render("Search")
	playlistTab := inactiveTabStyle.Render("Playlists")
	if m.tab == SearchTab {
		searchTab = activeTabStyle.Render("Search")
	} else {
		playlistTab = activeTabStyle.Render("Playlists")
	}
	title := titleStyle.Render(fmt.Sprintf("♪ findm - Music Finder (%s)", m.version))
	sb.WriteString(title + "  " + searchTab + " " + playlistTab + "\n\n")

	// Adding to playlist overlay
	if m.addingToPlaylist {
		if m.creatingPlaylist {
			sb.WriteString(titleStyle.Render("New playlist (track will be added automatically):") + "\n\n")
			sb.WriteString("Name: " + m.newPlaylistInput.View() + "\n")
			sb.WriteString(helpStyle.Render("Enter: create & add  Esc: back") + "\n")
			sb.WriteString("\n" + renderPlayerBar(m.player, m.width, m.animFrame) + "\n")
			return sb.String()
		}
		sb.WriteString(titleStyle.Render("Add to playlist:") + "\n\n")
		if len(m.playlists) == 0 {
			sb.WriteString(normalItemStyle.Render("No playlists yet. Press 'c' to create one.") + "\n")
		} else {
			helpBlock := helpStyle.Render("\nj/k: move  Enter: select  c: create new  Esc: cancel") + "\n"
			playerBlock := "\n" + renderPlayerBar(m.player, m.width, m.animFrame) + "\n"
			availableHeight := m.height - 2 - renderedLineCount(helpBlock) - renderedLineCount(playerBlock)
			if availableHeight < 1 {
				availableHeight = 1
			}
			visibleCount := fittedVisibleCount(len(m.playlists), m.addPlaylistCursor, availableHeight, func(start, end int) int {
				lines := end - start
				if start > 0 {
					lines++
				}
				if end < len(m.playlists) {
					lines++
				}
				return lines
			})
			start, end := viewportBounds(len(m.playlists), m.addPlaylistCursor, visibleCount)
			if start > 0 {
				sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↑ %d more", start)) + "\n")
			}
			for i := start; i < end; i++ {
				name := m.playlists[i]
				prefix := "  "
				style := normalItemStyle
				if i == m.addPlaylistCursor {
					prefix = "▸ "
					style = selectedItemStyle
				}
				sb.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, name)) + "\n")
			}
			if end < len(m.playlists) {
				sb.WriteString(scrollIndicatorStyle.Render(fmt.Sprintf("  ↓ %d more", len(m.playlists)-end)) + "\n")
			}
		}
		sb.WriteString(helpStyle.Render("\nj/k: move  Enter: select  c: create new  Esc: cancel") + "\n")
		sb.WriteString("\n" + renderPlayerBar(m.player, m.width, m.animFrame) + "\n")
		return sb.String()
	}

	// Creating playlist input
	if m.creatingPlaylist {
		sb.WriteString("New playlist name: " + m.newPlaylistInput.View() + "\n")
		sb.WriteString(helpStyle.Render("Enter: create  Esc: cancel") + "\n")
		sb.WriteString("\n" + renderPlayerBar(m.player, m.width, m.animFrame) + "\n")
		return sb.String()
	}

	statusBlock := ""
	if m.errorMsg != "" {
		statusBlock = "\n" + errorStyle.Render("Error: "+m.errorMsg) + "\n"
	} else if m.statusMsg != "" && (!m.loading || m.view != SearchView) {
		text := m.statusMsg
		if text == preparingPlaybackMsg {
			text = animatedLoadingMessage(text, m.animFrame)
		}
		statusBlock = "\n" + statusStyle.Render(text) + "\n"
	}

	playerBlock := "\n" + renderPlayerBar(m.player, m.width, m.animFrame) + "\n"
	visualizerBlock := ""
	if vizLine := renderVisualizer(m.viz, m.width); vizLine != "" {
		visualizerBlock = vizLine + "\n"
	}
	helpBlock := helpStyle.Render(helpText(m.view)) + "\n"

	availableHeight := m.height -
		2 -
		renderedLineCount(statusBlock) -
		renderedLineCount(playerBlock) -
		renderedLineCount(visualizerBlock) -
		renderedLineCount(helpBlock)
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Main content
	var content string
	switch m.view {
	case SearchView:
		content = "Search: " + m.searchInput.View()
		if m.loading {
			content += "\n\n" + statusStyle.Render("Searching...")
		}
	case ResultsView:
		content = renderResults(m.results, m.cursor, availableHeight, m.canLoadMoreResults())
	case PlaylistListView:
		content = renderPlaylistList(m.playlists, m.playlistCursor, availableHeight)
	case PlaylistDetailView:
		if m.currentPlaylist != nil {
			content = renderPlaylistDetail(m.currentPlaylist, m.trackCursor, availableHeight)
		}
	}
	sb.WriteString(strings.TrimRight(content, "\n"))

	sb.WriteString(statusBlock)
	sb.WriteString(playerBlock)
	sb.WriteString(visualizerBlock)
	sb.WriteString(helpBlock)

	return sb.String()
}

func helpText(view View) string {
	switch view {
	case SearchView:
		return "Enter: search (text, @handle, publisher, or URL)  Tab: next tab  Ctrl+C: quit"
	case ResultsView:
		return "j/k: move  Enter: play / open  n/p: next/prev  Space: pause  s: stop  h/l: seek  +/-: vol  r: recommend  a: add  Tab: next tab  /: search  q: quit"
	case PlaylistListView:
		return "j/k: move  Enter: open  c: create  d: delete  Tab: next tab  Esc: back  q: quit"
	case PlaylistDetailView:
		return "j/k: move  Enter: play  n/p: next/prev  Space: pause  s: stop  ←→: seek  +/-: vol  d: remove  Tab: next tab  Esc: back  q: quit"
	default:
		return ""
	}
}
