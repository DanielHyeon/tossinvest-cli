# Function Logic Map: `exposureReserve`

- Source: `internal/journal/reservations_test.go` (69-87)
- Qualified: `exposureReserve`
- AST evidence: `ast.json`(현재) — 현재 AST branches 0 · returns 1
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 83dc4732cbd12f85…) — 편집 전 갈래 수 0
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture — OPEN_EXPOSURE 예약 요청.

**a138 편집.** `OpenPositionLimit` 넉넉한 칸 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 직선 함수 — 분기 없음 | — | 정상 경로 | `go test ./internal/journal -run 'Reserv|A136|A138'` · `TestA136ReachingTheDailyBuyCeilingIsRefused` |

Exact AST return positions: 70

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 2종 — 대표: `a136ZeroManualBuy`, `j.Now`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
