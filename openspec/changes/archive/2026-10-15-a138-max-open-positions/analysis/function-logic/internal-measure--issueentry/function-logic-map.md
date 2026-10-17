# Function Logic Map: `issueEntry`

- Source: `internal/measure/reconstruct_test.go` (40-83)
- Qualified: `issueEntry`
- AST evidence: `ast.json`(현재) — 현재 AST branches 2 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` ae5c73dcf1be91b0…) — 편집 전 갈래 수 2
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture — 진입 발급.

**a138 편집.** `OpenPositionLimit` 넉넉한 칸 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 44) | — | 분기 본문 참조 | `go test ./internal/measure` |
| B2 | `if _, err := j.RecordDecisionAndReserve(ctx, journal.IssueRequest{` (if 48) | — | 분기 본문 참조 | `go test ./internal/measure` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 9종 — 대표: `a136ZeroManualBuy`, `context.Background`, `issued.Add`, `j.Now`, `j.RecordDecisionAndReserve`, `j.ReservationVersion`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
