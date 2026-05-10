# findm

![findm](findm.png)

YouTube 기반 터미널 음악 검색 및 재생 도구.

터미널에서 음악을 검색하고, 추천받고, 바로 재생할 수 있는
TUI 인터페이스를 제공합니다.

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

YouTube가 최근 영상 정보 추출 시 JavaScript 실행을 요구하는 흐름을
늘렸기 때문에, JS 런타임이 없으면 일부 영상(특히 Shorts와 최신
업로드)이 `This video is not available`로 떨어질 수 있습니다.
yt-dlp는 기본으로 `deno`를 인식하므로 설치만 해두면 됩니다.

```bash
# macOS
brew install deno

# 다른 플랫폼은 https://deno.com/ 참고
```

다른 JS 런타임을 쓰고 싶다면 yt-dlp에 명시할 수 있습니다.

```bash
yt-dlp --js-runtimes node ...
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
GitHub에서 태그를 push하면 GitHub Actions가
해당 태그명으로 자동 빌드하여 Release에 바이너리를 첨부합니다.

## 실행

```bash
./findm
```

## 데이터 저장 경로

| 파일 | 경로 | 설명 |
|------|------|------|
| 설정 파일 | `~/.config/findm/config.json` | API 키 등 설정 (선택, yt-dlp 사용 시 불필요) |
| 플레이리스트 | `~/.config/findm/playlists/*.json` | 저장된 플레이리스트 목록 |

설정 디렉토리는 `$XDG_CONFIG_HOME/findm/`을 따르며,
미설정 시 `~/.config/findm/`이 기본값입니다.
