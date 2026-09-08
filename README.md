# findm

![findm](findm.png)

A terminal music search and playback tool built on YouTube.

It provides a TUI for searching music, getting recommendations, and playing tracks right from the terminal.

## Prerequisites

### Install mpv

```bash
# macOS
brew install mpv

# Ubuntu/Debian
sudo apt install mpv

# Arch Linux
sudo pacman -S mpv
```

### Install yt-dlp

```bash
# macOS
brew install yt-dlp

# pip (all platforms)
pip install yt-dlp
```

### Install deno

YouTube increasingly requires JavaScript execution when extracting video info,
so without a JS runtime some videos (especially Shorts and recent uploads) may fail with `This video is not available`.
yt-dlp detects `deno` automatically, so installing it is enough.

```bash
# macOS
brew install deno
```

See: <https://github.com/yt-dlp/yt-dlp/wiki/EJS>

## Installation

```bash
go install github.com/ysoftman/findm@latest
```

Build from source:

```bash
go build -o findm .
```

`findm --version` (or `-v`) prints the version. It comes from the `-X main.Version=...` ldflag, or from the module version when installed with `go install ...@vX.Y.Z`, otherwise `dev`.
Pushing a tag to GitHub triggers GitHub Actions to build with that tag name and attach the binaries to the Release.

## Usage

```bash
./findm
```

## Thumbnails

In search results and playlist detail, the thumbnail of the video under the cursor is shown to the right of the list (terminal width of 90 columns or more).
It is fetched directly from `https://i.ytimg.com/vi/<ID>/mqdefault.jpg` without calling yt-dlp.

Terminals that support the Kitty graphics protocol (Ghostty, kitty) are detected automatically and get a high-resolution image;
everywhere else it is drawn with half-block (▀) truecolor characters.
Set `FINDM_THUMB=kitty` or `FINDM_THUMB=blocks` to force a mode.

To use Kitty mode inside tmux, tmux 3.3+ with `set -g allow-passthrough on` is required.

## Data Locations

| File | Path | Description |
|------|------|-------------|
| Config file | `~/.config/findm/config.json` | Settings such as API keys (optional, not needed with yt-dlp) |
| Playlists | `~/.config/findm/playlists/*.json` | Saved playlists |

The config directory follows `$XDG_CONFIG_HOME/findm/` and defaults to `~/.config/findm/` when unset.
