# Function Logic Map: `dispatchStrategyMarketHandoffs`

- Source: `internal/app/engine/strategy_market_handoff_delivery.go` (32-54, 편집 뒤)
- Qualified: `dispatchStrategyMarketHandoffs`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 44ed84069d5204e7…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` 26dda1be3cce793e…, 편집 전 분기수 3) — 현재 branches 4 · returns 5
- Risk scan: `risk-pattern-report.md`

**역할.** 한 시장 주기의 handoff 를 공유 dispatch 에 건네는 유일한 생산 자리.

**a137 편집.** 재리뷰 P1-N1: dispatch 오류가 `*execgw.DailyLossUnmeasurable` 면(B3 추가) ctx 기록자에 넘김 — 반환값 불변(전달 루프가 조용히 끝냄). 분기 3 → 4.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| ctx | 주기 ctx(기록자 실림) | runMarket | 기록자 없으면 무동작 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 38) | — | 분기 본문 참조 | 기존 dispatch 시험 |
| B2 | `if cas.Claimed \|\| cas.State != "FLAT" && cas.State != "CLOSED" {` (if 41) | — | 분기 본문 참조 | 기존 dispatch 시험 |
| B3 | `if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {` (if 46) | — | 분기 본문 참조 | `TestA137AnUnmeasurableDayIsCountedAndAnnouncedOncePerAccountDay`(effective · refreshOnly) |
| B4 | `if errors.Is(err, journal.ErrStrategyDispatchLeaseConsumed) {` (if 49) | — | 분기 본문 참조 | 기존 dispatch 시험 |

Exact AST return positions: 35 · 39 · 42 · 50 · 52

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `recordStrategyDailyLossUnmeasurable` | 계수 · 알림(a137) | — | AST calls |

## State mutations and fallbacks

- 없음(기록자 호출).

## Safety conclusion

- Safe edit boundary: 관측 한 줄.
- High-risk impact: yes(전략 주기) — 반환 · 분류 불변.
