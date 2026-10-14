# Function Logic Map: `TestProductionFirstLegAuthorityLoaderPairedKRUS`

- Source: `internal/app/engine/strategy_account_first_leg_authority_test.go` (20-85, 편집 뒤)
- Qualified: `TestProductionFirstLegAuthorityLoaderPairedKRUS`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 47f555f84382d276…, 편집 전 분기 수 11) — 현재 branches 11 · returns 0
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 fixture/시험(생산 코드 아님, tossos_testseams).

**a136 편집.** testseam fixture 의 Guardian 에 「수동 매수 0」 보드(a136 — nil 이면 진입 전부 거부). 시험이 재는 것 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| — | — | 시험 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 26) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B2 | `if !ok {` (if 30) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B3 | `if market == StrategyMarketUS {` (if 37) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B4 | `if market == StrategyMarketKR {` (if 48) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B5 | `} else {` (else 50) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B6 | `if err != nil {` (if 58) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B7 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 63) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B8 | `if refusal.Code != "" {` (if 66) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B9 | `if err != nil {` (if 70) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B10 | `if issuance.Entry.Market != string(market) \|\| issuance.Entry.EntryPrice != wantPrice \|\| issuance.Entry.QCandidate != result.Quantity \|\` (if 74) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B11 | `if !markets[StrategyMarketKR] \|\| !markets[StrategyMarketUS] {` (if 82) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |

Exact AST return positions: 

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no(시험 전용)
