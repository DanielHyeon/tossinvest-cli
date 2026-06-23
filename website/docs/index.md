# tossinvest-cli

**토스증권을 AI 에이전트·터미널에서 다루는 비공식 CLI.** 실행 바이너리는 `tossctl`.

토스증권 웹 세션을 재사용해 계좌·시세·거래내역 조회와 제한된 거래를 터미널에서 수행합니다.
Claude Code · Codex · Cursor · OpenClaw · bash · HTTP — 어떤 도구·에이전트든 동일한 명령
체계(`tossctl`)로 다룰 수 있고, 사람이 직접 터미널에서 쓸 수도 있습니다.

!!! warning "비공식 프로젝트"
    토스증권 공식 제품이 아닙니다. 웹 내부 API를 비공식적으로 사용하며, 이용약관(TOS)
    위반에 해당할 수 있습니다. API는 예고 없이 변경될 수 있고, 사용으로 인한 계좌 제한·손실·
    기타 불이익에 대해 개발자는 책임지지 않습니다. 본인의 판단과 책임 하에 사용하세요.

!!! danger "거래는 기본 비활성"
    설치 직후 모든 거래 기능이 꺼져 있습니다. `config.json`에서 기능별로 직접 허용해야만
    실행되며, 실거래는 `--execute` + `--confirm <token>` 2단계를 거칩니다. 자세한 내용은
    [안전 모델](safety.md) 참고.

## 핵심 요약

- **공식 Open API(예정)의 조회·거래 범위를 100% 커버**하고, 그 너머까지 다룹니다.
- 공식에 없는 기능: 수급(투자자별 순매수)·시장지수·지수 상세 시세·토스 AI 시그널·뉴스
  브리핑·조건검색(스크리너)·배당 내역·업종별 등락·커뮤니티 랭킹·관심종목 관리·거래내역
  ledger·실시간 푸시·소수점 주문·dry-run preview 등.
- 출력 형식: 표(table) · JSON · CSV · SSE — 자동화/에이전트 파이프라인에 바로 연결.

전체 비교는 [지원 범위](support-scope.md)를 참고하세요.

## 30초 시작

```bash
# 설치 (macOS / Linux)
curl -fsSL https://raw.githubusercontent.com/JungHoonGhae/tossinvest-cli/main/install.sh | sh

# 환경 점검 → 로그인 → 첫 조회
tossctl doctor
tossctl auth login
tossctl account summary --output json
```

자세한 설치는 [설치](install.md), 단계별 사용은 [빠른 시작](quickstart.md)을 보세요.

## AI 에이전트로 쓰는 경우

이 프로젝트는 **에이전트 친화적**으로 설계됐습니다. 모든 명령이 `--output json` 을
지원하고, 거래는 명시적 2단계 게이트로 보호됩니다. 에이전트 연동 지침은
[AI 에이전트 가이드](agents.md)에 정리돼 있습니다.

- **LLM용 문서**: [`/llms.txt`](https://junghoonghae.github.io/tossinvest-cli/llms.txt) (큐레이션 색인) · [`/llms-full.txt`](https://junghoonghae.github.io/tossinvest-cli/llms-full.txt) (전체 본문)
