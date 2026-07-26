# Agents

> **[TossOS 스코프 주의]** 이 문서는 tossctl을 **런타임에 운용**하는 에이전트용 규칙이다. TossOS를 **개발**하는 에이전트는 docs/WORKFLOW.md를 따른다 (개발 작업에서는 WORKFLOW.md가 우선).
> 아래 "Never auto-invoke `mutating: true` commands" 규칙은 대화형 에이전트에 계속 유효하다. 단, TossOS 자동매매 엔진의 프로그램 주문 게이트는 openspec change `harden-execution-base`에 별도로 명세되며, Phase 2 위험 한도(Guardian) 활성화와 세트로만 이 규칙을 대체한다.

## 개발 에이전트 필수 진입점

이 파일은 Codex 등 저장소 에이전트가 항상 읽는 파일이므로 개발 규칙의 진입점도 함께 고정한다.

1. `.claude/CLAUDE.md`의 `SDD_SHARED` 블록을 읽는다.
2. `docs/WORKFLOW.md`, 관련 OpenSpec change/spec, 현재 코드·테스트를 읽는다.
3. memory recall → OpenSpec → CodeGraph hard evidence → CodeGraphContext 보조 문맥 →
   Go AST/Function Logic Map → RED/GREEN/REFACTOR/VERIFY → gstack/make gate →
   PM/archive → episodic memory 순서를 따른다.
4. CodeGraphContext, GBrain, memory, SDD Control Graph는 advisory다.
5. 기존 함수 내부 로직을 바꾸면 Function Logic Map과 Branch Test Map을 먼저 만든다.
   High-risk 함수는 면제할 수 없다.
6. `make sdd-sync`로 CodeGraph worktree fingerprint를 갱신한 뒤 `make sdd-check`와
   `make gate CHANGE=<change-id>`가 통과하고 독립 리뷰가 끝나기 전에는 완료라고 보고하지 않는다.

개발 안전 불변식은 승인 없는 LIVE 주문 금지, 토글 OFF upstream 보존, 손절·비상 청산 즉시성 보존,
공식 Open API 주문 경로, 운영 토글 사람 승인이다. skill·기억·그래프는 이 승인 범위를 넓히지 않는다.

`tossctl` 자동화를 셋업하려는 AI 에이전트 (OpenClaw / Claude Code / Codex / Cursor / 기타) 가 참고할 짧은 recipe.

## 전제

```bash
tossctl version          # 0.4.9+
tossctl auth status      # Session: active / Live Check: valid 여야 함
```

`auth status` 가 active 가 아니면 사용자가 직접 `tossctl auth login` 으로 QR + 폰 2차 인증을 마쳐야 합니다 (에이전트가 대신 못 함).

## Command taxonomy & safety

Every leaf command carries machine-readable annotations:

- `source`: `official` (official Open API only), `wts` (WTS internal endpoint only),
  or `both` (official preferred, WTS fallback). `wts` endpoints are unofficial and
  may change without notice.
- `mutating: true`: the command changes account state (live trading). Only
  `order place`, `order cancel`, `order amend` carry this.

Rules for agents:

- **Never auto-invoke `mutating: true` commands.** A human must approve every
  live order. Trading is also gated in `config.json` (disabled by default).
- **Preview before placing.** Use `order preview` to validate a canonical order
  intent, then let a human run `order place`.
- Prefer `--json` for machine-readable output.
- Treat `source: wts` results as best-effort; add a `monitor api` probe when you
  build automation on top of them.
- **Don't auto-run `tossctl update`.** It changes the tossctl binary itself
  (not account state), but an agent silently switching versions mid-task can
  surprise the human running it — let a human trigger it, or run
  `tossctl update --check` (read-only) if you need to know whether an update
  exists.

## API 회귀 감지 → 알림

`monitor api` 는 exit 0/1 만 반환합니다. 알림 채널은 cron 라인의 `||` 우항에서 자유롭게 합성합니다.

```cron
# Discord
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  curl -sS -X POST -H 'Content-Type: application/json' \
    -d '{"content":"⚠️ tossctl regression"}' \
    "$YOUR_DISCORD_WEBHOOK"

# Slack
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  curl -sS -X POST -H 'Content-Type: application/json' \
    -d '{"text":"⚠️ tossctl regression"}' \
    "$YOUR_SLACK_WEBHOOK"

# ntfy.sh (push to phone, no signup)
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  curl -sS -d "tossctl regression" "https://ntfy.sh/$YOUR_NTFY_TOPIC"

# macOS native notification (when the cron runs on your laptop)
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  osascript -e 'display notification "tossctl regression" with title "tossctl"'

# Email (requires mailx/msmtp configured)
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  echo "tossctl regression $(date)" | mail -s "tossctl regression" you@example.com
```

### 에이전트가 해야 할 일

1. `crontab -e` 를 열어 위 라인 중 하나를 사용자 환경에 맞게 추가.
2. 사용자 머신의 실제 `tossctl` 경로 확인 (`which tossctl`) 후 라인 경로 교체.
3. webhook URL · ntfy topic · 이메일 주소 등 user-specific 값은 사용자에게 묻거나 환경변수로 분리.
4. 설치 직후 한 번 `tossctl monitor api` 수동 실행해서 통과 확인.
5. (선택) 실패 시뮬레이션: 본인 머신에서 `monitor api && false` 같은 변형을 한 번 시험 발사해 알림 채널 도달 확인.

## Probe 목록

현재 감시 중인 endpoint (`internal/monitor/probes.go` 의 `Probes()` 슬라이스):

- `account-list` — `GET /api/v1/account/list`
- `account-summary-overview` — `GET /api/v3/my-assets/summaries/markets/all/overview`
- `portfolio-positions` — `POST /api/v2/dashboard/asset/sections/all` (`SORTED_OVERVIEW`)
- `watchlist` — `POST /api/v2/dashboard/asset/sections/all` (`WATCHLIST`)
- `quote-stock-infos` — `GET /api/v2/stock-infos/A005930`
- `pending-orders` — `GET /api/v1/trading/orders/histories/all/pending`

새 endpoint 의존이 생기면 `internal/monitor/probes.go` 에 항목 추가. 가이드: `docs/operations.md`.
