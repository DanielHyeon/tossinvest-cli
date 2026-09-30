# Function Logic Map: `RecordOnly.Notify`

- Source: `internal/obs/record_only.go` (`45`–`61`)
- Qualified: `RecordOnly.Notify`
- AST evidence: `ast.json` (`source_sha256` 465f094668570d3a…) — base `b30318d6` 에서 `go run ./tools/logic-map`
- Risk scan: `risk-pattern-report.md`
- 분기 3 · 반환 4

**편집.** 편집하지 않는다 — 새 critical 종류의 기록 · 로그 경로 증거(D5 · H2).

**역할.** 사건을 로그 한 줄로 남기고(`logEvent` — 종류 = `e.Type`), critical 이면 원장에 기록만, 일반이면 유계 버퍼에 넘긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.N` | nil 허용 | 배선 | nil → 무동작(B1) |
| `e.Type` | 종류 | 호출자 | `SeverityOf` 로 등급 |

## Branches and early returns

> 조건은 소스 원문, 진입 실측은 base `b30318d6` 의 `go test -count=1 -coverprofile`(covermode set, 2026-10-01)에서 그 줄로 시작하는 블록의 count (`analysis/harness/write_bundles.py`). engine 패키지 실행은 `-trimpath` 로 `TestA111…` 두 시험이 소스 경로를 못 찾아 실패했다 — 커버리지 프로파일은 그대로 쓰인다(두 시험은 이 함수들과 무관한 AST 핀).

| Branch | 종류 | 조건 (원문) | 진입 실측 |
|---|---|---|---|
| B1 | if | `:47` `if n == nil {` | 예 |
| B2 | if | `:52` `if severity != SeverityCritical {` | 예 |
| B3 | if | `:53` `if r.Relay == nil {` | 예 |

Exact AST return positions: `48:3`, `55:4`, `58:3`, `60:2`

## Calls and live bindings

| Callee | Line | Why called | Error/timeout/retry contract |
|---|---|---|---|
| `SeverityOf` | `:50` | 등급 | 순수(맵 조회) |
| `n.logEvent` | `:51` | 구조화 로그 한 줄 | 로컬 로그 — **알림과 같은 종류**(H2 근거: 알림 경로 자체는 이미 한 종류) |
| `withoutFields` | `:51` | 필드 제거(계좌 원문 차단) | 순수 |
| `n.logNormalDrop` | `:54` | 릴레이 없음 → 버림 기록 | 로그 한 줄 |
| `r.Relay.Offer` | `:57` | 일반 등급 이관 | 비차단(`normal_relay.go:53-57`) |
| `n.recordCritical` | `:60` | critical 기록 | `n.mu` + `Journal.RecordAlert`(busy_timeout 5s, 원격 0), 실패 → 게이트 래치 + 승격 시도 |
| `n.remindAfter` | `:60` | 재알림 창 | `DefaultRemindAfter` 1h |

## State mutations and fallbacks

outbox 행 삽입 또는 재무장(critical) · 버퍼 적재(normal).

## Safety conclusion

- 새 종류가 critical 로 등록되면 이 함수의 B2(normal 갈래)를 건너 `recordCritical` 로 간다 — 그 사실 하나가 a091 의 이익(원장 흔적)이다.
