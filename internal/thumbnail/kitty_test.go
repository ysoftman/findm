package thumbnail

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"math/rand"
	"strings"
	"testing"
)

const gridStart = "\x1b[38;5;1m"

// splitKitty separates the transmission chunks (without their ST) from the
// placeholder grid.
func splitKitty(t *testing.T, out string) (chunks []string, grid string) {
	t.Helper()
	gi := strings.Index(out, gridStart)
	if gi < 0 {
		t.Fatalf("no placeholder grid in %q", out)
	}
	for c := range strings.SplitSeq(out[:gi], "\x1b\\") {
		if c != "" {
			chunks = append(chunks, c)
		}
	}
	return chunks, out[gi:]
}

func decodeChunks(t *testing.T, chunks []string) image.Image {
	t.Helper()
	var payload strings.Builder
	for _, c := range chunks {
		_, p, ok := strings.Cut(c, ";")
		if !ok {
			t.Fatalf("chunk without payload: %q", c)
		}
		payload.WriteString(p)
	}
	raw, err := base64.StdEncoding.DecodeString(payload.String())
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	return img
}

func TestRenderKitty(t *testing.T) {
	out := renderKitty(testImage(), 4, 1, false)
	if !strings.HasPrefix(out, "\x1b_Ga=T,U=1,q=2,f=100,t=d,i=1,c=4,r=1,m=0;") {
		t.Fatalf("unexpected prefix: %q", out[:60])
	}
	chunks, grid := splitKitty(t, out)
	if b := decodeChunks(t, chunks).Bounds(); b.Dx() != 4 || b.Dy() != 2 {
		t.Fatalf("decoded bounds = %v, want 4x2", b)
	}
	if strings.Contains(grid, "\n") {
		t.Fatalf("grid has more than 1 line: %q", grid)
	}
	if n := strings.Count(grid, "\U0010EEEE"); n != 4 {
		t.Fatalf("placeholders = %d, want 4", n)
	}
	cells := []rune(strings.TrimSuffix(strings.TrimPrefix(grid, gridStart), "\x1b[39m"))
	if len(cells) != 12 {
		t.Fatalf("grid runes = %d, want 12", len(cells))
	}
	if cells[1] != diacritics[0] || cells[2] != diacritics[0] {
		t.Fatalf("first cell diacritics = %U %U, want row 0 col 0", cells[1], cells[2])
	}
	if cells[10] != diacritics[0] || cells[11] != diacritics[3] {
		t.Fatalf("last cell diacritics = %U %U, want row 0 col 3", cells[10], cells[11])
	}
}

func TestRenderKittyChunks(t *testing.T) {
	// Random noise defeats PNG compression so the payload spans several chunks.
	rnd := rand.New(rand.NewSource(1))
	img := image.NewRGBA(image.Rect(0, 0, 64, 64))
	for i := range img.Pix {
		img.Pix[i] = uint8(rnd.Intn(256))
	}
	chunks, _ := splitKitty(t, renderKitty(img, 8, 4, false))
	if len(chunks) < 2 {
		t.Fatalf("chunks = %d, want several", len(chunks))
	}
	for i, c := range chunks[:len(chunks)-1] {
		if !strings.Contains(c, "m=1;") || (i > 0 && !strings.HasPrefix(c, "\x1b_Gm=1;")) {
			t.Fatalf("chunk %d not marked continued: %q", i, c[:20])
		}
	}
	if last := chunks[len(chunks)-1]; !strings.HasPrefix(last, "\x1b_Gm=0;") {
		t.Fatalf("last chunk not final: %q", last[:20])
	}
	if b := decodeChunks(t, chunks).Bounds(); b.Dx() != 64 || b.Dy() != 64 {
		t.Fatalf("decoded bounds = %v, want 64x64", b)
	}
}

func TestRenderKittyTmux(t *testing.T) {
	img := testImage()
	plain := renderKitty(img, 4, 1, false)
	out := renderKitty(img, 4, 1, true)
	if !strings.HasPrefix(out, "\x1bPtmux;\x1b\x1b_G") {
		t.Fatalf("unexpected prefix: %q", out[:20])
	}
	gi := strings.Index(out, gridStart)
	var unwrapped strings.Builder
	for w := range strings.SplitSeq(out[:gi], "\x1bPtmux;") {
		if w != "" {
			unwrapped.WriteString(strings.ReplaceAll(strings.TrimSuffix(w, "\x1b\\"), "\x1b\x1b", "\x1b"))
		}
	}
	if got := unwrapped.String() + out[gi:]; got != plain {
		t.Fatalf("unwrapped tmux output differs from plain output")
	}
}

func TestRenderKittyFallback(t *testing.T) {
	img := testImage()
	if got := renderKitty(img, len(diacritics)+1, 1, false); got != renderBlocks(img, len(diacritics)+1, 1) {
		t.Fatal("oversized grid should fall back to half-blocks")
	}
}

func TestDetectKitty(t *testing.T) {
	t.Setenv("FINDM_THUMB", "kitty")
	if !detectKitty() {
		t.Fatal("FINDM_THUMB=kitty should enable kitty")
	}
	t.Setenv("FINDM_THUMB", "blocks")
	if detectKitty() {
		t.Fatal("FINDM_THUMB=blocks should disable kitty")
	}
}
