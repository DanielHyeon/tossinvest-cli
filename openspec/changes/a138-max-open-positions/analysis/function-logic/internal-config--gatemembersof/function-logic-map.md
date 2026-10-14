# Function Logic Map: `gateMembersOf`

- Source: `internal/config/limits_io.go` (192-214)
- Qualified: `gateMembersOf`
- AST evidence: `ast.json`(현재) — 현재 AST branches 3 · returns 2
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6c8fdc5d37028679…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 편집기 저장이 건드리는 키 목록.

**a138 편집.** `OpenPositions != nil` 이면 `max_open_positions` 멤버 — 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 194) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B2 | `if l.DailyBuy != nil {` (if 205) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |
| B3 | `if l.OpenPositions != nil {` (if 209) | — | 분기 본문 참조 | `go test ./internal/config -run 'Limits|Save|A138'` · `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` |

Exact AST return positions: 195 · 213

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 3종 — 대표: `append`, `formatLimit`, `json.Marshal`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
