package thumbnail

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

// testImage is 4x2: top row red, bottom row blue.
func testImage() *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, 4, 2))
	for x := range 4 {
		img.Set(x, 0, color.RGBA{255, 0, 0, 255})
		img.Set(x, 1, color.RGBA{0, 0, 255, 255})
	}
	return img
}

func TestURL(t *testing.T) {
	got := URL("abc123")
	want := "https://i.ytimg.com/vi/abc123/mqdefault.jpg"
	if got != want {
		t.Fatalf("URL() = %q, want %q", got, want)
	}
}

func TestRenderBlocks(t *testing.T) {
	out := renderBlocks(testImage(), 4, 1)
	if lines := strings.Count(out, "\n") + 1; lines != 1 {
		t.Fatalf("lines = %d, want 1", lines)
	}
	if n := strings.Count(out, "▀"); n != 4 {
		t.Fatalf("cells = %d, want 4", n)
	}
	if !strings.HasPrefix(out, "\x1b[38;2;255;0;0m\x1b[48;2;0;0;255m▀") {
		t.Fatalf("unexpected first cell: %q", out)
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Fatalf("missing reset at end: %q", out)
	}
}
