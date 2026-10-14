# Function Logic Map: `seedExpiredUnconsumedReservation`

- Source: `internal/app/engine/engine_test.go` (486-536, 편집 뒤)
- Qualified: `seedExpiredUnconsumedReservation`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 2ece46493d087d62…, 편집 전 분기 수 4) — 현재 branches 4 · returns 1
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험 함수(생산 코드 아님).

**a136 편집.** 시험 fixture 에 a136 의 새 필수 칸을 채움 — EXPOSURE_RAISING 예약에 넉넉한 `DailyBuyLimit`(이 시험이 재는 것과 무관)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 495) | — | 분기 본문 참조 | `go test ./internal/app/engine` (이 함수를 쓰는 기존 시험) |
| B2 | `if err != nil {` (if 511) | — | 분기 본문 참조 | `go test ./internal/app/engine` (이 함수를 쓰는 기존 시험) |
| B3 | `if err != nil {` (if 515) | — | 분기 본문 참조 | `go test ./internal/app/engine` (이 함수를 쓰는 기존 시험) |
| B4 | `if _, err := j.Reserve(ctx, journal.ReserveRequest{` (if 519) | — | 분기 본문 참조 | `go test ./internal/app/engine` (이 함수를 쓰는 기존 시험) |

Exact AST return positions: 535

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture 구성 | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 추가 — 시험이 단언하는 값 무변경.
- High-risk impact: no(시험 전용, 비례 원칙)
