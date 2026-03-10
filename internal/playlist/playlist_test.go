package playlist

import (
	"testing"
)

func TestPlaylistCRUD(t *testing.T) {
	name := "test-crud-playlist"

	// Cleanup first
	_ = Delete(name)

	// Create
	if err := Create(name); err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// List
	names, err := List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	found := false
	for _, n := range names {
		if n == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("playlist %q not found in list: %v", name, names)
	}

	// Add track
	track := Track{
		VideoID: "abc123",
		Title:   "Test Song",
		Channel: "Test Channel",
		URL:     "https://www.youtube.com/watch?v=abc123",
	}
	if err := AddTrack(name, track); err != nil {
		t.Fatalf("AddTrack failed: %v", err)
	}

	// Get and verify
	pl, err := Get(name)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(pl.Tracks) != 1 {
		t.Fatalf("expected 1 track, got %d", len(pl.Tracks))
	}
	if pl.Tracks[0].VideoID != "abc123" {
		t.Fatalf("expected videoID abc123, got %s", pl.Tracks[0].VideoID)
	}

	// Remove track
	if err := RemoveTrack(name, 0); err != nil {
		t.Fatalf("RemoveTrack failed: %v", err)
	}
	pl, err = Get(name)
	if err != nil {
		t.Fatalf("Get after remove failed: %v", err)
	}
	if len(pl.Tracks) != 0 {
		t.Fatalf("expected 0 tracks after remove, got %d", len(pl.Tracks))
	}

	// Delete
	if err := Delete(name); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}
