# Function Logic Map: `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange`

- Source: `internal/execgw/a137_daily_loss_internal_test.go` (287-341, 편집 뒤)
- Qualified: `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 486545ca9d8ea620…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` bca3eba9aebc69d6…, 편집 전 분기수 8) — 현재 branches 10 · returns 4
- Risk scan: `risk-pattern-report.md`

**역할.** 제출 직전 펜스의 일일 손실 표지 분류 시험.

**a137 편집.** 구현 리뷰 P2-1: reconcile latch 를 먼저 · 일일 손실 투영을 뒤에 두는 경우(gw4) 추가 — 분기 8 → 10.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 관문 | 시험 EntryGate | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 291) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B2 | `if !a137IsAccountBlock(err) {` (if 296) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B3 | `if fenced.Reason != ReasonStrategyDispatchFenced \|\| !a137IsAccountBlock(fenced) {` (if 300) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B4 | `if err != nil {` (if 307) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B5 | `if err := gw3.withStrategyEntryGateAuthority(ctx, expected3, "kr", "005930", func() error { return nil }); err == nil \|\| a137IsAccountBlock(err) {` (if 312) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B6 | `if err != nil {` (if 319) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B7 | `if err := gw4.withStrategyEntryGateAuthority(ctx, expected4, "kr", "005930", func() error { return nil }); err == nil \|\| a137IsAccountBlock(err) {` (if 324) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B8 | `if err != nil {` (if 330) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B9 | `if err == nil \|\| a137IsAccountBlock(err) {` (if 335) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |
| B10 | `if fenced := strategyFenceRejection(err); a137IsAccountBlock(fenced) \|\| fenced.Reason != ReasonStrategyDispatchFenced {` (if 338) | — | 분기 본문 참조 | `go test ./internal/execgw -run TestA137TheFinalGate` |

Exact AST return positions: 295 · 312 · 324 · 334

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `withStrategyEntryGateAuthority` | 펜스 | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 시험.
- High-risk impact: no — 시험.
