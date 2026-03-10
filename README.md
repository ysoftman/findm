# findm

YouTube 기반 터미널 음악 검색 및 재생 도구.

터미널에서 음악을 검색하고, 추천받고, 바로 재생할 수 있는
TUI 인터페이스를 제공합니다.

## 주요 기능

- **키워드 검색** - YouTube 음악 키워드 검색
  (예: "카페 음악", "한국 인디", "acoustic chill")
- **유사 음악 추천** - 선택한 곡 기반 관련 음악 추천
- **오디오 재생** - mpv를 통한 오디오 전용 재생
- **로컬 플레이리스트** - JSON 파일로 플레이리스트 관리
- **뷰포트 스크롤링** - 긴 목록 스크롤 + 인디케이터
- **동적 레이아웃** - 터미널 너비에 자동 적응

## 사전 요구 사항

- [Go](https://go.dev/) 1.21+
- [mpv](https://mpv.io/) - 오디오 재생 엔진
- [yt-dlp](https://github.com/yt-dlp/yt-dlp) - YouTube 검색 및 메타데이터 추출

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

## 설치

```bash
go install github.com/ysoftman/findm@latest
```

소스에서 빌드:

```bash
git clone https://github.com/ysoftman/findm.git
cd findm
go build -o findm .
```

## 실행

```bash
./findm
```

## 키 바인딩

### 공통

| 키 | 동작 |
|----|------|
| `Tab` | 검색 / 플레이리스트 탭 전환 |
| `/` | 검색창 열기 |
| `Ctrl+C` | 종료 |
| `q` | 종료 (검색 입력 중 제외) |

### 검색 / 결과

| 키 | 동작 |
|----|------|
| `Enter` | 검색 실행 / 선택한 곡 재생 |
| `j` / `k` | 커서 아래 / 위 이동 |
| `r` | 선택한 곡의 유사 음악 추천 |
| `a` | 선택한 곡을 플레이리스트에 추가 |
| `Space` | 재생 중지 |
| `Esc` | 검색 입력에서 결과로 돌아가기 |

### 플레이리스트

| 키 | 동작 |
|----|------|
| `Enter` | 플레이리스트 열기 / 곡 재생 |
| `j` / `k` | 커서 아래 / 위 이동 |
| `c` | 새 플레이리스트 생성 |
| `d` | 플레이리스트 삭제 / 곡 제거 |
| `Space` | 재생 중지 |
| `Esc` | 뒤로 가기 |

## 프로젝트 구조

```text
findm/
├── main.go                  # 엔트리포인트
├── internal/
│   ├── config/config.go     # 설정 관리
│   ├── youtube/
│   │   ├── client.go        # yt-dlp 기반 클라이언트
│   │   ├── search.go        # 음악 검색
│   │   └── recommend.go     # 유사 음악 추천
│   ├── player/player.go     # mpv 오디오 재생
│   ├── playlist/playlist.go # 로컬 플레이리스트 CRUD
│   └── tui/
│       ├── app.go           # bubbletea 메인 모델
│       ├── search.go        # 검색 뷰
│       ├── results.go       # 결과 리스트 (뷰포트)
│       ├── player.go        # 플레이어 바
│       ├── playlist.go      # 플레이리스트 뷰
│       └── styles.go        # lipgloss 스타일
└── go.mod
```

## 데이터 저장 경로

- 설정: `~/.config/findm/config.json`
- 플레이리스트: `~/.config/findm/playlists/*.json`

## 라이선스

MIT
