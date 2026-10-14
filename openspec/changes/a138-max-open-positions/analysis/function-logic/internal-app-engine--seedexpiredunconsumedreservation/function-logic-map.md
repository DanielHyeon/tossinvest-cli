# Function Logic Map: `seedExpiredUnconsumedReservation`

- Source: `internal/app/engine/engine_test.go` (486-537)
- Qualified: `seedExpiredUnconsumedReservation`
- AST evidence: `ast.json`(현재) — 현재 AST branches 4 · returns 1
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 92ff107c69a985f0…) — 편집 전 갈래 수 4
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture — 예약 요청.

**a138 편집.** `OpenPositionLimit` 넉넉한 칸 하나(이 시험이 재는 것과 무관).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 495) | — | 분기 본문 참조 | `go test ./internal/app/engine` |
| B2 | `if err != nil {` (if 511) | — | 분기 본문 참조 | `go test ./internal/app/engine` |
| B3 | `if err != nil {` (if 515) | — | 분기 본문 참조 | `go test ./internal/app/engine` |
| B4 | `if _, err := j.Reserve(ctx, journal.ReserveRequest{` (if 519) | — | 분기 본문 참조 | `go test ./internal/app/engine` |

Exact AST return positions: 536

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 16종 — 대표: `Add`, `UTC`, `a136ZeroManualBuy`, `clock.NewFake`, `context.Background`, `filepath.Join`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
