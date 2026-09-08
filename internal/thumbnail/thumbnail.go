// Package thumbnail fetches YouTube video thumbnails and renders them for
// terminal display, via the Kitty graphics protocol when supported and
// half-block truecolor ANSI text otherwise.
package thumbnail

import (
	"fmt"
	"image"
	_ "image/jpeg"
	"net/http"
	"os"
	"strings"
	"time"
)

// URL returns the 320x180 "mqdefault" thumbnail URL for a YouTube video ID.
func URL(id string) string {
	return "https://i.ytimg.com/vi/" + id + "/mqdefault.jpg"
}

// Fetch downloads and decodes the thumbnail for a YouTube video ID.
func Fetch(id string) (image.Image, error) {
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(URL(id))
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("thumbnail %s: status %d", id, resp.StatusCode)
	}
	img, _, err := image.Decode(resp.Body)
	return img, err
}

// Render draws img cols cells wide, keeping its aspect ratio at two pixel
// rows per cell row. It uses Kitty graphics placeholders when the terminal
// supports them and half-block (▀) truecolor ANSI text otherwise.
func Render(img image.Image, cols int) string {
	b := img.Bounds()
	rows := max(cols*b.Dy()/b.Dx()/2, 1)
	if useKitty() {
		return renderKitty(img, cols, rows, os.Getenv("TMUX") != "")
	}
	return renderBlocks(img, cols, rows)
}

// renderBlocks draws img as half-block (▀) truecolor ANSI text. Each cell
// packs two vertically adjacent sample pixels: the upper one as the
// foreground color, the lower one as the background color.
func renderBlocks(img image.Image, cols, rows int) string {
	var sb strings.Builder
	for r := range rows {
		if r > 0 {
			sb.WriteByte('\n')
		}
		for c := range cols {
			tr, tg, tb := avg(img, c, r*2, cols, rows*2)
			br, bg, bb := avg(img, c, r*2+1, cols, rows*2)
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", tr, tg, tb, br, bg, bb)
		}
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

// avg returns the mean color of the source pixels covered by cell column cx,
// pixel-row py in a cols x prows sampling grid.
func avg(img image.Image, cx, py, cols, prows int) (uint8, uint8, uint8) {
	b := img.Bounds()
	x0 := b.Min.X + cx*b.Dx()/cols
	x1 := b.Min.X + (cx+1)*b.Dx()/cols
	y0 := b.Min.Y + py*b.Dy()/prows
	y1 := b.Min.Y + (py+1)*b.Dy()/prows
	var r, g, bl, n uint64
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			cr, cg, cb, _ := img.At(x, y).RGBA()
			r += uint64(cr >> 8)
			g += uint64(cg >> 8)
			bl += uint64(cb >> 8)
			n++
		}
	}
	if n == 0 {
		return 0, 0, 0
	}
	return uint8(r / n), uint8(g / n), uint8(bl / n)
}
