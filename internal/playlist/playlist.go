package playlist

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ysoftman/findm/internal/config"
)

// Track represents a single track in a playlist.
type Track struct {
	VideoID string `json:"video_id"`
	Title   string `json:"title"`
	Channel string `json:"channel"`
	URL     string `json:"url"`
}

// Playlist represents a named collection of tracks.
type Playlist struct {
	Name   string  `json:"name"`
	Tracks []Track `json:"tracks"`
}

// playlistDir returns the playlists directory path.
func playlistDir() (string, error) {
	dir, err := config.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "playlists"), nil
}

// playlistPath returns the file path for a named playlist.
func playlistPath(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" || trimmed == "." || trimmed == ".." {
		return "", fmt.Errorf("invalid playlist name: %q", name)
	}
	dir, err := playlistDir()
	if err != nil {
		return "", err
	}
	safeName := strings.ReplaceAll(name, "/", "_")
	return filepath.Join(dir, safeName+".json"), nil
}

// Create creates a new empty playlist.
func Create(name string) error {
	path, err := playlistPath(name)
	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("playlist %q already exists", name)
	}

	pl := &Playlist{Name: name, Tracks: []Track{}}
	return save(path, pl)
}

// Delete removes a playlist.
func Delete(name string) error {
	path, err := playlistPath(name)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil {
		return fmt.Errorf("failed to delete playlist %q: %w", name, err)
	}
	return nil
}

// List returns all playlist names.
func List() ([]string, error) {
	dir, err := playlistDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read playlists directory: %w", err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, strings.TrimSuffix(e.Name(), ".json"))
		}
	}
	return names, nil
}

// Get loads a playlist by name.
func Get(name string) (*Playlist, error) {
	path, err := playlistPath(name)
	if err != nil {
		return nil, err
	}
	return load(path)
}

// AddTrack adds a track to a playlist.
func AddTrack(name string, track Track) error {
	path, err := playlistPath(name)
	if err != nil {
		return err
	}

	pl, err := load(path)
	if err != nil {
		return err
	}

	// Check for duplicates
	for _, t := range pl.Tracks {
		if t.VideoID == track.VideoID {
			return fmt.Errorf("track %q already in playlist", track.Title)
		}
	}

	pl.Tracks = append(pl.Tracks, track)
	return save(path, pl)
}

// RemoveTrack removes a track from a playlist by index.
func RemoveTrack(name string, index int) error {
	path, err := playlistPath(name)
	if err != nil {
		return err
	}

	pl, err := load(path)
	if err != nil {
		return err
	}

	if index < 0 || index >= len(pl.Tracks) {
		return fmt.Errorf("track index %d out of range", index)
	}

	pl.Tracks = append(pl.Tracks[:index], pl.Tracks[index+1:]...)
	return save(path, pl)
}

func load(path string) (*Playlist, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load playlist: %w", err)
	}

	var pl Playlist
	if err := json.Unmarshal(data, &pl); err != nil {
		return nil, fmt.Errorf("failed to parse playlist: %w", err)
	}
	return &pl, nil
}

func save(path string, pl *Playlist) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create playlists directory: %w", err)
	}

	data, err := json.MarshalIndent(pl, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal playlist: %w", err)
	}

	return os.WriteFile(path, data, 0o644)
}
