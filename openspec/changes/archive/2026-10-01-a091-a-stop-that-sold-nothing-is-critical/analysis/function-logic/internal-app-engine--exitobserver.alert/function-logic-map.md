# Function Logic Map: `ExitObserver.alert`

- Source: `internal/app/engine/exitloop.go` (`1832`–`1839`)
- Qualified: `ExitObserver.alert`
- AST evidence: `ast.json` (`source_sha256` 0733bd8641ed8c41…) — 편집 뒤 `540aebe6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 2 · 반환 1

**편집.** 편집하지 않는다 — 부분 캡 알림이 타는 기존 경로. **a091 의 0주 보고는 이 함수를 거치지 않는다**(`notifyZeroFloor` 가 `o.opts.Alerts.Notify(context.WithoutCancel)` 를 직접 부름 — 이 함수의 실패 로그 `logErr` 가 계좌 원문을 싣기 때문, design D8 · 변이 M15).

**역할.** exit 관측 goroutine 의 알림 입구. 오류는 로그 한 줄로 삼킨다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `o.opts.Alerts` | nil 허용 | 생산: `obs.RecordOnly`(`exitwiring.go:349`) | nil 이면 무동작(B1) |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 편집 뒤 `540aebe6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:1833` `if o.opts.Alerts == nil {` | 예 |
| B2 | if | `:1836` `if err := o.opts.Alerts.Notify(ctx, e); err != nil {` | 예 |

Exact AST return positions: `1834:3`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `o.opts.Alerts.Notify` | `:1836` | 기록(critical) 또는 이관(normal) | `o.opts.Alerts` = `obs.RecordOnly{N, Relay}`(생산 배선 `exitwiring.go:348-349`). 일반 등급 → `NormalRelay.Offer`(`normal_relay.go:43-58` — 비차단 `select`, 버퍼가 차면 버림을 로그로 기록). critical → `n.mu` 아래 `Journal.RecordAlert`(`record_only.go:136-137` — SQLite `BEGIN IMMEDIATE`, `busy_timeout` **5s**(`journal.go:36`) · `synchronous=FULL` fsync, 원격 전송 0). `n.mu` 대기는 **기한 없음**(보유자는 로컬 원장 연산만 — 운영자 승인 `Acknowledge` 는 밀린 행 수에 비례). 재알림 창 `DefaultRemindAfter` **1h**(`notifier.go:59`). 기록 실패 → 그 자리에서 게이트 래치 + 승격(`escalate` 원장 트랜잭션, 동기) 후 오류 반환. 실패 오류는 이 함수가 `logErr` 로 삼킨다 |
| `o.logErr` | `:1837` | 기록 실패 로그 | 로그 한 줄 — 종류는 사건의 종류(`e.Type`), **계좌 필드 원문**(그래서 a091 은 이 경로를 쓰지 않는다) |

## State mutations and fallbacks

원장 outbox 행(critical) 또는 이관 버퍼(normal). 기록 실패는 알림기가 이미 게이트를 잠갔다.

## Safety conclusion

- a091 의 새 종류는 같은 알림기(`RecordOnly`)를 쓰되 이 함수를 우회한다(D8). 원격 전송 0.
