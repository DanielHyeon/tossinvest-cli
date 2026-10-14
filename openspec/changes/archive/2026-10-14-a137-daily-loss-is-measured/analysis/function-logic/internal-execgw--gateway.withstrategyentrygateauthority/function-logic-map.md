# Function Logic Map: `Gateway.withStrategyEntryGateAuthority`

- Source: `internal/execgw/strategy_entry_gate_authority.go` (46-81, 편집 뒤)
- Qualified: `Gateway.withStrategyEntryGateAuthority`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 0d91fbdd3bcf5df5…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 6d2de28f31db7a3c…, 편집 전 분기수 6) — 현재 branches 8 · returns 8
- Risk scan: `risk-pattern-report.md`

**역할.** 제출 직전 최종 관문(세대 · latch 재확인 아래 fn 실행).

**a137 편집.** 세대가 바뀌었거나(B4) 계좌 latch 거절(B6)이고 지금 상태가 AUTO 일일 손실 모드만이면 표지를 단 모드 거절을 돌려줌(B5 · B7 추가). 그 밖은 오늘과 같음. 분기 6 → 8.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| expected | 관측 때 봉인한 세대 | ObserveStrategyEntryGate | 다르면 B4 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| g.entry == nil \|\| fn == nil \|\| expected.generation == 0 \|\| expected.digest == "" {` (if 49) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` |
| B2 | `if refresh != nil {` (if 56) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` |
| B3 | `if err := refresh(); err != nil {` (if 57) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` |
| B4 | `if current != expected {` (if 64) | — | 분기 본문 참조 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` |
| B5 | `if marked := g.entry.dailyLossOnlyRejectionLocked(market, symbol); marked != nil {` (if 66) | — | 분기 본문 참조 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` · `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock` |
| B6 | `if rejected := g.entry.checkAccountEntryLocked(); rejected != nil {` (if 71) | — | 분기 본문 참조 | `TestA137OtherFencesBehaveAsBefore` · 기존 `TestStrategyGatewayRechecksEntryGateAfterSubmittingFencePairedKRUS` |
| B7 | `if marked := g.entry.dailyLossOnlyRejectionLocked(market, symbol); marked != nil && marked.Detail == rejected.Detail {` (if 72) | — | 분기 본문 참조 | `TestA137TheFinalGateFenceMarksOnlyTheDailyLossChange` |
| B8 | `if rejected := g.entry.checkSymbolEntryLocked(market, symbol); rejected != nil {` (if 77) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run StrategyGateway` |

Exact AST return positions: 50 · 58 · 67 · 69 · 73 · 75 · 78 · 80

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.entry.dailyLossOnlyRejectionLocked` | 표지(a137, g.mu 아래) | 아니면 nil | AST calls |
| `fn` | 브로커 호출 | 관문 잠금 아래 | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 거절 값의 표지만.
- High-risk impact: yes(제출 직전 관문) — 거절 여부 무변경, 브로커 요청 0 유지.
