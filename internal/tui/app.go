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
	"github.com/ysoftman/findm/internal/youtube"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// handlePlayerErr handles player errors, showing ErrNotReady as status instead of error.
func (m *Model) handlePlayerErr(err error) {
	if err == nil {
		return
	}
	if errors.Is(err, player.ErrNotReady) {
		m.statusMsg = "Preparing playback..."
	} else {
		m.errorMsg = err.Error()
	}
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
	client *youtube.Client
	player *player.Player

	// UI state
	tab         Tab
	view        View
	searchInput textinput.Model
	width       int
	height      int

	// Search results
	results []youtube.Video
	cursor  int

	// Playlist state
	playlists       []string
	playlistCursor  int
	currentPlaylist *playlist.Playlist
	trackCursor     int

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
func NewModel(client *youtube.Client) Model {
	si := newSearchInput()
	si.Focus()

	pi := textinput.New()
	pi.Placeholder = "Playlist name..."
	pi.CharLimit = 50
	pi.Width = 40

	return Model{
		client:           client,
		player:           player.New(),
		tab:              SearchTab,
		view:             SearchView,
		searchInput:      si,
		newPlaylistInput: pi,
	}
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
		m.statusMsg = fmt.Sprintf("Found %d results", len(msg.results))
		return m, nil

	case recommendMsg:
		m.loading = false
		if msg.err != nil {
			m.errorMsg = msg.err.Error()
			return m, nil
		}
		m.results = msg.results
		m.cursor = 0
		m.view = ResultsView
		m.statusMsg = fmt.Sprintf("Found %d recommendations", len(msg.results))
		return m, nil

	case tickMsg:
		state := m.player.GetState()
		if state != player.Stopped {
			if state == player.Preparing {
				m.statusMsg = "Preparing playback..."
			} else {
				if m.statusMsg == "Preparing playback..." {
					m.statusMsg = ""
				}
			}
			return m, tickCmd()
		}
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
		m.loading = true
		m.statusMsg = "Searching..."
		m.searchInput.Blur()
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

func (m Model) handleResultsView(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.cursor < len(m.results)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "enter":
		if len(m.results) > 0 {
			v := m.results[m.cursor]
			if err := m.player.Play(v.URL, v.Title); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = ""
				return m, tickCmd()
			}
		}
	case " ":
		m.handlePlayerErr(m.player.TogglePause())
	case "s":
		m.player.Stop()
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
		if len(m.results) > 0 {
			v := m.results[m.cursor]
			m.loading = true
			m.statusMsg = "Finding recommendations..."
			return m, performRecommend(m.client, v.ID)
		}
	case "a":
		if len(m.results) > 0 {
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
			t := m.currentPlaylist.Tracks[m.trackCursor]
			if err := m.player.Play(t.URL, t.Title); err != nil {
				m.errorMsg = err.Error()
			} else {
				m.statusMsg = ""
				return m, tickCmd()
			}
		}
	case " ":
		m.handlePlayerErr(m.player.TogglePause())
	case "s":
		m.player.Stop()
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

	// Title
	sb.WriteString(titleStyle.Render("♪ findm - Music Finder") + "\n\n")

	// Tabs
	searchTab := inactiveTabStyle.Render("Search")
	playlistTab := inactiveTabStyle.Render("Playlists")
	if m.tab == SearchTab {
		searchTab = activeTabStyle.Render("Search")
	} else {
		playlistTab = activeTabStyle.Render("Playlists")
	}
	sb.WriteString(searchTab + " " + playlistTab + "\n\n")

	// Adding to playlist overlay
	if m.addingToPlaylist {
		if m.creatingPlaylist {
			sb.WriteString(titleStyle.Render("New playlist (track will be added automatically):") + "\n\n")
			sb.WriteString("Name: " + m.newPlaylistInput.View() + "\n")
			sb.WriteString(helpStyle.Render("Enter: create & add  Esc: back") + "\n")
			sb.WriteString("\n" + renderPlayerBar(m.player, m.width) + "\n")
			return sb.String()
		}
		sb.WriteString(titleStyle.Render("Add to playlist:") + "\n\n")
		if len(m.playlists) == 0 {
			sb.WriteString(normalItemStyle.Render("No playlists yet. Press 'c' to create one.") + "\n")
		} else {
			for i, name := range m.playlists {
				prefix := "  "
				style := normalItemStyle
				if i == m.addPlaylistCursor {
					prefix = "▸ "
					style = selectedItemStyle
				}
				sb.WriteString(style.Render(fmt.Sprintf("%s%s", prefix, name)) + "\n")
			}
		}
		sb.WriteString(helpStyle.Render("\nj/k: move  Enter: select  c: create new  Esc: cancel") + "\n")
		sb.WriteString("\n" + renderPlayerBar(m.player, m.width) + "\n")
		return sb.String()
	}

	// Creating playlist input
	if m.creatingPlaylist {
		sb.WriteString("New playlist name: " + m.newPlaylistInput.View() + "\n")
		sb.WriteString(helpStyle.Render("Enter: create  Esc: cancel") + "\n")
		sb.WriteString("\n" + renderPlayerBar(m.player, m.width) + "\n")
		return sb.String()
	}

	// Calculate available height for list content
	// Reserve lines: title(2) + tabs(2) + status(2) + player(2) + help(2) = ~10
	availableHeight := m.height - 10
	if availableHeight < 3 {
		availableHeight = 3
	}

	// Main content
	switch m.view {
	case SearchView:
		sb.WriteString("Search: " + m.searchInput.View() + "\n")
		if m.loading {
			sb.WriteString("\n" + statusStyle.Render("Searching...") + "\n")
		}
	case ResultsView:
		sb.WriteString(renderResults(m.results, m.cursor, availableHeight) + "\n")
	case PlaylistListView:
		sb.WriteString(renderPlaylistList(m.playlists, m.playlistCursor, availableHeight) + "\n")
	case PlaylistDetailView:
		if m.currentPlaylist != nil {
			sb.WriteString(renderPlaylistDetail(m.currentPlaylist, m.trackCursor, availableHeight) + "\n")
		}
	}

	// Status/Error (skip statusMsg when loading to avoid duplicate)
	if m.errorMsg != "" {
		sb.WriteString("\n" + errorStyle.Render("Error: "+m.errorMsg) + "\n")
	} else if m.statusMsg != "" && (!m.loading || m.view != SearchView) {
		sb.WriteString("\n" + statusStyle.Render(m.statusMsg) + "\n")
	}

	// Player bar
	sb.WriteString("\n" + renderPlayerBar(m.player, m.width) + "\n")

	// Help
	var help string
	switch m.view {
	case SearchView:
		help = "Enter: search  Tab: playlists  Ctrl+C: quit"
	case ResultsView:
		help = "j/k: move  Enter: play  Space: pause  s: stop  h/l: seek  +/-: vol  r: recommend  a: add  /: search  q: quit"
	case PlaylistListView:
		help = "j/k: move  Enter: open  c: create  d: delete  Tab: search  Esc: back  q: quit"
	case PlaylistDetailView:
		help = "j/k: move  Enter: play  Space: pause  s: stop  ←→: seek  +/-: vol  d: remove  Esc: back  q: quit"
	}
	sb.WriteString(helpStyle.Render(help) + "\n")

	return sb.String()
}
