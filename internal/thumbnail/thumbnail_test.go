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
	// 2x3 samples per cell: the top two sample rows are red, the bottom
	// one blue, so the red upper four cells form SEXTANT-1234 (U+1FB0E).
	if n := strings.Count(out, "\U0001FB0E"); n != 4 {
		t.Fatalf("cells = %d, want 4", n)
	}
	if !strings.HasPrefix(out, "\x1b[38;2;255;0;0m\x1b[48;2;0;0;255m\U0001FB0E") {
		t.Fatalf("unexpected first cell: %q", out)
	}
	if !strings.HasSuffix(out, "\x1b[0m") {
		t.Fatalf("missing reset at end: %q", out)
	}
}

func TestSextant(t *testing.T) {
	cases := map[int]rune{
		0: ' ', 0b010101: '▌', 0b101010: '▐', 0b111111: '█',
		1: 0x1FB00, 0b001111: 0x1FB0E, 0b010110: 0x1FB14, 0b101011: 0x1FB28, 0b111110: 0x1FB3B,
	}
	for pat, want := range cases {
		if got := sextant(pat); got != want {
			t.Errorf("sextant(%06b) = %U, want %U", pat, got, want)
		}
	}
}

func TestSplitUniform(t *testing.T) {
	var px [6][3]int
	for i := range px {
		px[i] = [3]int{10, 20, 30}
	}
	fg, bg, pat := split(px)
	if pat != 0 || fg != bg || bg != [3]int{10, 20, 30} {
		t.Fatalf("split(uniform) = %v %v %06b", fg, bg, pat)
	}
}
