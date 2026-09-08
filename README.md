# findm

![findm](findm.png)

YouTube 기반 터미널 음악 검색 및 재생 도구.

터미널에서 음악을 검색하고, 추천받고, 바로 재생할 수 있는 TUI 인터페이스를 제공합니다.

## 사전 요구 사항

### mpv 설치

```bash
# macOS
brew install mpv

# Ubuntu/Debian
sudo apt install mpv

# Arch Linux
sudo pacman -S mpv
```

### yt-dlp 설치

```bash
# macOS
brew install yt-dlp

# pip (모든 플랫폼)
pip install yt-dlp
```

### deno 설치

YouTube가 최근 영상 정보 추출 시 JavaScript 실행을 요구하는 흐름을 늘렸기 때문에,
JS 런타임이 없으면 일부 영상(특히 Shorts와 최신 업로드)이 `This video is not available`로 떨어질 수 있습니다.
yt-dlp는 기본으로 `deno`를 인식하므로 설치만 해두면 됩니다.

```bash
# macOS
brew install deno
```

참고: <https://github.com/yt-dlp/yt-dlp/wiki/EJS>

## 설치

```bash
go install github.com/ysoftman/findm@latest
```

소스에서 빌드:

```bash
go build -o findm .
```

버전을 지정하지 않으면 타이틀에 `dev`로 표시됩니다.
GitHub에서 태그를 push하면 GitHub Actions가 해당 태그명으로 자동 빌드하여 Release에 바이너리를 첨부합니다.

## 실행

```bash
./findm
```

## 썸네일

검색 결과와 플레이리스트 상세에서 커서 위 영상의 썸네일을 목록 오른쪽에 표시합니다(터미널 폭 90열 이상).
`https://i.ytimg.com/vi/<ID>/mqdefault.jpg`를 직접 받아오며 yt-dlp 호출은 없습니다.

Kitty graphics protocol을 지원하는 터미널(Ghostty, kitty)은 자동 감지해 고해상도로 그리고,
그 외에는 하프블록(▀) truecolor 문자로 그립니다.
`FINDM_THUMB=kitty` 또는 `FINDM_THUMB=blocks`로 강제 지정할 수 있습니다.

tmux 안에서 Kitty 모드를 쓰려면 tmux 3.3+ 에서 `set -g allow-passthrough on` 설정이 필요합니다.

## 데이터 저장 경로

| 파일 | 경로 | 설명 |
|------|------|------|
| 설정 파일 | `~/.config/findm/config.json` | API 키 등 설정 (선택, yt-dlp 사용 시 불필요) |
| 플레이리스트 | `~/.config/findm/playlists/*.json` | 저장된 플레이리스트 목록 |

설정 디렉토리는 `$XDG_CONFIG_HOME/findm/`을 따르며,
미설정 시 `~/.config/findm/`이 기본값입니다.
