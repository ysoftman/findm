// Package thumbnail fetches YouTube video thumbnails and renders them for
// terminal display, via the Kitty graphics protocol when supported and
// sextant (2x3 block) truecolor ANSI text otherwise.
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
// supports them and sextant (2x3 block) truecolor ANSI text otherwise.
func Render(img image.Image, cols int) string {
	b := img.Bounds()
	rows := max(cols*b.Dy()/b.Dx()/2, 1)
	if useKitty() {
		return renderKitty(img, cols, rows, os.Getenv("TMUX") != "", directPlace())
	}
	return renderBlocks(img, cols, rows)
}

// renderBlocks draws img as sextant (2x3 block) truecolor ANSI text. Each
// cell samples six pixels, splits them into two color groups and draws the
// high group as the foreground sextant pattern over the low group's color.
func renderBlocks(img image.Image, cols, rows int) string {
	var sb strings.Builder
	for r := range rows {
		if r > 0 {
			sb.WriteByte('\n')
		}
		for c := range cols {
			var px [6][3]int
			for i := range px {
				px[i] = avg(img, c*2+i%2, r*3+i/2, cols*2, rows*3)
			}
			fg, bg, pat := split(px)
			fmt.Fprintf(&sb, "\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm%c",
				fg[0], fg[1], fg[2], bg[0], bg[1], bg[2], sextant(pat))
		}
		sb.WriteString("\x1b[0m")
	}
	return sb.String()
}

// split partitions six sample colors at the midpoint of the channel with the
// widest range. It returns the mean color of the high side (fg) and the low
// side (bg) and the bit pattern of the high side, bit i standing for sample i.
// Uniform samples all land on the low side, giving pat 0 and fg == bg.
func split(px [6][3]int) (fg, bg [3]int, pat int) {
	ch, lo, hi := 0, 0, -1
	for k := range 3 {
		l, h := 255, 0
		for _, p := range px {
			l, h = min(l, p[k]), max(h, p[k])
		}
		if h-l > hi-lo {
			ch, lo, hi = k, l, h
		}
	}
	mid := (lo + hi) / 2
	var fsum, bsum [3]int
	fn, bn := 0, 0
	for i, p := range px {
		if p[ch] > mid {
			pat |= 1 << i
			fn++
			for k := range 3 {
				fsum[k] += p[k]
			}
		} else {
			bn++
			for k := range 3 {
				bsum[k] += p[k]
			}
		}
	}
	for k := range 3 {
		bg[k] = bsum[k] / bn
	}
	if fn == 0 {
		return bg, bg, 0
	}
	for k := range 3 {
		fg[k] = fsum[k] / fn
	}
	return fg, bg, pat
}

// sextant returns the block character whose filled cells match pat, bit i
// standing for cell i in reading order (0 top-left .. 5 bottom-right). The
// Symbols for Legacy Computing block omits the four patterns that already
// exist as Block Elements.
func sextant(pat int) rune {
	switch pat {
	case 0:
		return ' '
	case 0b010101:
		return '▌'
	case 0b101010:
		return '▐'
	case 0b111111:
		return '█'
	}
	r := rune(0x1FB00 + pat - 1)
	if pat > 0b010101 {
		r--
	}
	if pat > 0b101010 {
		r--
	}
	return r
}

// avg returns the mean color of the source pixels covered by cell column cx,
// pixel-row py in a cols x prows sampling grid. A grid finer than the image
// still covers at least one pixel per cell.
func avg(img image.Image, cx, py, cols, prows int) [3]int {
	b := img.Bounds()
	x0 := b.Min.X + cx*b.Dx()/cols
	x1 := max(b.Min.X+(cx+1)*b.Dx()/cols, x0+1)
	y0 := b.Min.Y + py*b.Dy()/prows
	y1 := max(b.Min.Y+(py+1)*b.Dy()/prows, y0+1)
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
	return [3]int{int(r / n), int(g / n), int(bl / n)}
}
