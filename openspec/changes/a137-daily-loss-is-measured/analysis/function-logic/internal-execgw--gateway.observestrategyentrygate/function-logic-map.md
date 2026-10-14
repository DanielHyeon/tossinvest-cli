# Function Logic Map: `Gateway.ObserveStrategyEntryGate`

- Source: `internal/execgw/strategy_entry_gate_authority.go` (22-35, 편집 뒤)
- Qualified: `Gateway.ObserveStrategyEntryGate`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 0d91fbdd3bcf5df5…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 6d2de28f31db7a3c…, 편집 전 분기수 2) — 현재 branches 2 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 진입 관문 권한 관측(거절 또는 봉인된 세대).

**a137 편집.** 거절을 그대로 돌리는 대신 `g.entry.markStrategyRefusal(refusal, market, symbol)` — AUTO 일일 손실 모드 latch 하나뿐이고 종목 latch 도 없으며 문구가 같을 때만 표지(cause)를 단 사본. 분기 수 불변(2).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| market · symbol | kr/us · 비지 않음 | 호출자 | B1 거절 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil \|\| g.entry == nil \|\| ctx == nil \|\| ctx.Err() != nil \|\| (market != "kr" && market != "us") \|\| symbol == "" {` (if 24) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/execgw -run StrategyEntryGate` |
| B2 | `if refusal := g.entry.CheckEntryFor(market, symbol); refusal != nil {` (if 27) | — | 분기 본문 참조 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(양방향 6 경우) |

Exact AST return positions: 25 · 29 · 34

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.entry.CheckEntryFor` | 관문 판정 | 거절 → B2 | AST calls |
| `g.entry.markStrategyRefusal` | 표지(a137) | 그 밖은 그대로 | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 반환 오류의 표지만. 사유 · 문구 무변경.
- High-risk impact: yes(진입 관문 분류) — 차단 여부 무변경.
