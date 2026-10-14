# Function Logic Map: `EngineBrokerOrderIDs`

- Source: `internal/journal/daily_buy.go` (350-371, 편집 뒤)
- Qualified: `Journal.EngineBrokerOrderIDs`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` c5577a1ada9550a2…, 편집 전 분기 수 4) — 현재 branches 4 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 엔진 주문 id 집합(수동 매수 중복 제거).

**a136 편집.** 844a042c: PLACE · AMEND attempt 만(구현 리뷰 P2-2) — 질의 조건 하나, 분기 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 355) | — | 분기 본문 참조 | `TestA136EngineOrderIDsExcludeCancelTargets` · `TestA136TheRefresherPublishesTodaysManualBuys` |
| B2 | `for rows.Next() {` (for 360) | — | 분기 본문 참조 | `TestA136EngineOrderIDsExcludeCancelTargets` · `TestA136TheRefresherPublishesTodaysManualBuys` |
| B3 | `if err := rows.Scan(&id); err != nil {` (if 362) | — | 분기 본문 참조 | `TestA136EngineOrderIDsExcludeCancelTargets` · `TestA136TheRefresherPublishesTodaysManualBuys` |
| B4 | `if err := rows.Err(); err != nil {` (if 367) | — | 분기 본문 참조 | `TestA136EngineOrderIDsExcludeCancelTargets` · `TestA136TheRefresherPublishesTodaysManualBuys` |

Exact AST return positions: 356 · 363 · 368 · 370

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | — | — | AST |

## State mutations and fallbacks

- 편집 내용 참조.

## Safety conclusion

- Safe edit boundary: 844a042c: PLACE · AMEND attempt 만(구현 리뷰 P2-2) — 질의 조건 하나, 분기 무변경.
- High-risk impact: yes(합계 입력 — 잘못 빼면 과소 집계)
