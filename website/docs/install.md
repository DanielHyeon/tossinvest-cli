# 설치

`tossctl` 은 단일 실행 바이너리입니다. macOS·Linux·Windows를 지원합니다.

## macOS / Linux (한 줄 설치)

```bash
curl -fsSL https://raw.githubusercontent.com/JungHoonGhae/tossinvest-cli/main/install.sh | sh
```

다운로드·체크섬 검증·PATH 등록까지 자동으로 처리합니다.

## Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/JungHoonGhae/tossinvest-cli/main/install.ps1 | iex
```

## Homebrew

```bash
brew install JungHoonGhae/tossinvest-cli/tossctl
# 업그레이드
brew upgrade tossctl
```

## 소스 빌드 (Go 1.25+)

```bash
go install github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl@latest
```

## 사전 요구사항

- **로그인(`auth login`)에는 Google Chrome 과 Python 이 필요**합니다. 한 줄 설치
  스크립트가 자동으로 설정하며, 수동 환경에서는 `tossctl doctor` 가 무엇이 빠졌는지
  알려줍니다.

## 설치 확인

```bash
tossctl version
tossctl doctor
```

`doctor` 는 바이너리·Chrome·Python·세션·엔드포인트 상태를 점검합니다. 문제가 있으면
`tossctl doctor --report` 로 JSON 진단(홈 경로 자동 마스킹)을 만들어 이슈에 붙일 수 있습니다.
