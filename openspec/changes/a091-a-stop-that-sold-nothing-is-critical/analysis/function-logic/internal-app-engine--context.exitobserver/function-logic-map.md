# Function Logic Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (`319`–`360`)
- Qualified: `Context.ExitObserver`
- AST evidence: `ast.json` (`source_sha256` 27905e84de89b7d9…) — 편집 뒤 `3ec1efd2` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 6 · 반환 4

**편집.** **a091 편집(구현 로트)** — 로드된 설정 `c.Config.Engine.Notifications.Enabled` 로 `opts.NotificationsEnabled` 를 덮는 한 줄(`:355`). 분기 무변경(변이 M5 CAUGHT).

**역할.** exit 관측 루프의 생산 조립 — 알림 · 통지 · 조회 경로를 주입 지점별로 기록 전용으로 덮고, 하한 공급자를 exit Retrier 로 만든다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `c.Config.Engine.Notifications.Enabled` | 로드된 설정 | 설정 파일 | 거짓이 기본 — 보호 0주가 옛 종류로 남는다 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `3ec1efd2` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:320` `if c == nil {` | 아니오 |
| B2 | if | `:323` `if !c.Automation.Verified {` | 예 |
| B3 | if | `:327` `if !ok {` | 아니오 |
| B4 | if | `:343` `if opts.Names == nil {` | 예 |
| B5 | if | `:348` `if c.Notifier != nil {` | 예 |
| B6 | if | `:356` `if opts.Floor == nil {` | 예 |

Exact AST return positions: `321:3`, `324:3`, `328:3`, `359:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `fmt.Errorf` | `:324` | 미검증 조립 거절 | 순수 |
| `fmt.Errorf` | `:328` | Guardian 형 거절 | 순수 |
| `c.NormalAlertRelay` | `:333` | 일반 등급 이관 버퍼 | 메모리 |
| `exitSideRetrier` | `:334` | exit 전용 Retrier 사본 | 메모리 |
| `exitSideFloor` | `:357` | exit 전용 하한 공급자 | 메모리 |
| `NewExitObserver` | `:359` | 루프 생성 | 검증 |

## State mutations and fallbacks

없음(옵션 조립).

## Safety conclusion

- 덮기는 호출자 값과 무관하다 — a095 `reconcileloop.go:369` 와 같은 원천 · 같은 규칙. High-risk: yes(생산 배선).
