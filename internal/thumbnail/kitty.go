package thumbnail

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// useKitty reports whether the terminal supports Kitty graphics Unicode placeholders.
var useKitty = sync.OnceValue(detectKitty)

// detectKitty honors FINDM_THUMB=kitty|blocks, otherwise sniffs the terminal
// (through tmux's global environment when running inside tmux).
func detectKitty() bool {
	switch os.Getenv("FINDM_THUMB") {
	case "kitty":
		return true
	case "blocks":
		return false
	}
	prog := os.Getenv("TERM_PROGRAM")
	if os.Getenv("TMUX") != "" {
		if out, err := exec.Command("tmux", "show-environment", "-g", "TERM_PROGRAM").Output(); err == nil {
			prog = strings.TrimSpace(strings.TrimPrefix(string(out), "TERM_PROGRAM="))
		}
	}
	return prog == "ghostty" || prog == "kitty" || os.Getenv("KITTY_WINDOW_ID") != ""
}

// renderKitty transmits img as PNG via the Kitty graphics protocol (image id 1)
// and returns a cols x rows grid of Unicode placeholder cells the terminal
// fills in. The transmission escapes are zero-width and precede the first row.
// When tmux is true each chunk is wrapped in a DCS passthrough.
func renderKitty(img image.Image, cols, rows int, tmux bool) string {
	if cols > len(diacritics) || rows > len(diacritics) {
		return renderBlocks(img, cols, rows)
	}
	// png.Encode emits 16-bit truecolor for non-RGBA sources such as a decoded
	// JPEG's YCbCr; converting first roughly halves the payload.
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, img.Bounds().Min, draw.Src)
	var buf bytes.Buffer
	if err := png.Encode(&buf, rgba); err != nil {
		return renderBlocks(img, cols, rows)
	}
	data := base64.StdEncoding.EncodeToString(buf.Bytes())
	var sb strings.Builder
	for i := 0; i < len(data); i += 4096 {
		end := min(i+4096, len(data))
		ctrl := "m=1"
		if end == len(data) {
			ctrl = "m=0"
		}
		if i == 0 {
			ctrl = fmt.Sprintf("a=T,U=1,q=2,f=100,t=d,i=1,c=%d,r=%d,%s", cols, rows, ctrl)
		}
		chunk := "\x1b_G" + ctrl + ";" + data[i:end] + "\x1b\\"
		if tmux {
			chunk = "\x1bPtmux;" + strings.ReplaceAll(chunk, "\x1b", "\x1b\x1b") + "\x1b\\"
		}
		sb.WriteString(chunk)
	}
	for r := range rows {
		if r > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString("\x1b[38;5;1m")
		for c := range cols {
			sb.WriteRune('\U0010EEEE')
			sb.WriteRune(diacritics[r])
			sb.WriteRune(diacritics[c])
		}
		sb.WriteString("\x1b[39m")
	}
	return sb.String()
}

// diacritics are the first 64 row/column markers from kitty's
// rowcolumn-diacritics.txt, indexed by row or column number.
var diacritics = []rune{
	0x0305, 0x030D, 0x030E, 0x0310, 0x0312, 0x033D, 0x033E, 0x033F,
	0x0346, 0x034A, 0x034B, 0x034C, 0x0350, 0x0351, 0x0352, 0x0357,
	0x035B, 0x0363, 0x0364, 0x0365, 0x0366, 0x0367, 0x0368, 0x0369,
	0x036A, 0x036B, 0x036C, 0x036D, 0x036E, 0x036F, 0x0483, 0x0484,
	0x0485, 0x0486, 0x0487, 0x0592, 0x0593, 0x0594, 0x0595, 0x0597,
	0x0598, 0x0599, 0x059C, 0x059D, 0x059E, 0x059F, 0x05A0, 0x05A1,
	0x05A8, 0x05A9, 0x05AB, 0x05AC, 0x05AF, 0x05C4, 0x0610, 0x0611,
	0x0612, 0x0613, 0x0614, 0x0615, 0x0616, 0x0617, 0x0657, 0x0658,
}
