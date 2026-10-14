# Function Logic Map: `newA112TradingFixture`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (83-166, 편집 뒤)
- Qualified: `newA112TradingFixture`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 3aa6cc58f8149bcc…, 편집 전 분기 수 13) — 현재 branches 13 · returns 3
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
| B1 | `if symbols == nil {` (if 86) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B2 | `for _, market := range []StrategyMarket{StrategyMarketKR, StrategyMarketUS} {` (range 94) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B3 | `if !ok {` (if 98) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B4 | `if market == StrategyMarketKR {` (if 104) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B5 | `} else {` (else 106) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B6 | `if options.secondLane != nil {` (if 112) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B7 | `if options.coordinatorOrder {` (if 119) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B8 | `if err != nil {` (if 125) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B9 | `if options.maxOpenExposure != "" {` (if 131) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B10 | `if options.riskBudget != "" {` (if 134) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B11 | `if err != nil {` (if 139) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B12 | `if options.disallowFor != "" {` (if 150) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |
| B13 | `if err != nil \|\| config.Symbol != options.disallowFor {` (if 154) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine` |

Exact AST return positions: 155 · 159 · 165

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | fixture | — | AST |

## State mutations and fallbacks

- 시험 데이터만.

## Safety conclusion

- Safe edit boundary: 칸 하나.
- High-risk impact: no(시험 전용)
