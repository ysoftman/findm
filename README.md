# findm

![findm](findm.png)

YouTube 기반 터미널 음악 검색 및 재생 도구.

터미널에서 음악을 검색하고, 추천받고, 바로 재생할 수 있는
TUI 인터페이스를 제공합니다.

## 주요 기능

- **키워드 검색** - YouTube 음악 키워드 검색
  (예: "카페 음악", "한국 인디", "acoustic chill")
- **유사 음악 추천** - 선택한 곡 기반 관련 음악 추천
- **오디오 재생** - mpv IPC 소켓을 통한 오디오 재생 (일시정지, 탐색, 볼륨 조절)
- **프로그레스 바** - 실시간 재생 위치, 전체 시간, 볼륨 표시
- **오디오 비주얼라이저** - 실제 오디오에 반응하는 visualizer 표시
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

버전을 지정하여 빌드:

```bash
go build -ldflags "-X main.Version=1.0.0" -o findm .
```

버전을 지정하지 않으면 타이틀에 `dev`로 표시됩니다.
GitHub에서 태그를 push하면 GitHub Actions가
해당 태그명으로 자동 빌드하여 Release에 바이너리를 첨부합니다.

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
| `Esc` | 검색 입력에서 결과로 돌아가기 |

### 재생 제어 (결과 / 플레이리스트 뷰)

| 키 | 동작 |
|----|------|
| `Space` | 일시정지 / 재개 |
| `s` | 재생 정지 |
| `h` / `←` | 10초 뒤로 탐색 |
| `l` / `→` | 10초 앞으로 탐색 |
| `+` / `=` | 볼륨 증가 |
| `-` | 볼륨 감소 |

### 플레이리스트

| 키 | 동작 |
|----|------|
| `Enter` | 플레이리스트 열기 / 곡 재생 |
| `j` / `k` | 커서 아래 / 위 이동 |
| `c` | 새 플레이리스트 생성 |
| `d` | 플레이리스트 삭제 / 곡 제거 |
| `Esc` | 뒤로 가기 |

## 프로젝트 구조

```text
findm/
├── main.go                  # 엔트리포인트
├── version.go               # 버전 정보 (ldflags로 주입)
├── internal/
│   ├── config/config.go     # 설정 관리
│   ├── youtube/
│   │   ├── client.go        # yt-dlp 기반 클라이언트
│   │   ├── search.go        # 음악 검색
│   │   └── recommend.go     # 유사 음악 추천
│   ├── player/player.go     # mpv IPC 소켓 기반 오디오 재생
│   ├── playlist/playlist.go # 로컬 플레이리스트 CRUD
│   ├── visualizer/
│   │   └── visualizer.go    # 오디오 비주얼라이저 애니메이션
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

| 파일 | 경로 | 설명 |
|------|------|------|
| 설정 파일 | `~/.config/findm/config.json` | API 키 등 설정 (선택, yt-dlp 사용 시 불필요) |
| 플레이리스트 | `~/.config/findm/playlists/*.json` | 저장된 플레이리스트 목록 |

설정 디렉토리는 `$XDG_CONFIG_HOME/findm/`을 따르며,
미설정 시 `~/.config/findm/`이 기본값입니다.

## 라이선스

MIT
